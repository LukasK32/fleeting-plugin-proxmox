package plugin

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/luthermonson/go-proxmox"
	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
	"golang.org/x/sync/errgroup"
)

var _ provider.InstanceGroup = (*InstanceGroup)(nil)

const triggerChannelCapacity = 100

var ErrNoIPv4Address = errors.New("failed to determine IPv4 address for instance")

type InstanceGroup struct {
	Settings         `json:",inline"`
	FleetingSettings provider.Settings `json:"-"`

	log     hclog.Logger    `json:"-"`
	proxmox *proxmox.Client `json:"-"`

	// This mutex is used when clonning template for new instances. It is required for blocking other
	// operations like collection or update, because when new instance is created with recycled ID then for
	// a brief period it will be reported from Proxmox with old name (e.g. InstanceNameRemoving).
	instanceClonningMu sync.Mutex `json:"-"`

	// Trigger for collector to start removed instances collection.
	instanceCollectionTrigger chan struct{} `json:"-"`

	// Trigger to shutdown collector.
	collectorShutdownTrigger chan struct{} `json:"-"`

	// Wait group for the collector.
	collectorWaitGroup sync.WaitGroup `json:"-"`
}

// Init implements provider.InstanceGroup.
func (ig *InstanceGroup) Init(ctx context.Context, logger hclog.Logger, settings provider.Settings) (provider.ProviderInfo, error) {
	var err error

	ig.log = logger
	ig.FleetingSettings = settings
	ig.instanceCollectionTrigger = make(chan struct{}, triggerChannelCapacity)
	ig.collectorShutdownTrigger = make(chan struct{}, 1)

	if err := ig.Settings.CheckRequiredFields(); err != nil {
		return provider.ProviderInfo{}, err
	}

	ig.Settings.FillWithDefaults()

	if ig.Settings.InsecureSkipTLSVerify {
		ig.log.Warn("TLS verification for Proxmox client is disabled, connections will be insecure")
	}

	ig.proxmox, err = ig.getProxmoxClient()
	if err != nil {
		return provider.ProviderInfo{}, err
	}

	if err := ig.markStaleInstancesForRemoval(ctx); err != nil {
		return provider.ProviderInfo{}, err
	}

	// Sleep for a bit to give Proxmox a chance to propagate renames for stale instances
	// Without this sleep these instances would be reported as creating during first Update
	<-time.After(collectionWaitAfterTrigger)

	//nolint:contextcheck
	ig.startRemovedInstanceCollector()

	return provider.ProviderInfo{
		ID:      ig.Settings.Pool,
		MaxSize: *ig.Settings.MaxInstances,
	}, nil
}

// Shutdown implements provider.InstanceGroup.
func (ig *InstanceGroup) Shutdown(_ context.Context) error {
	ig.collectorShutdownTrigger <- struct{}{}
	ig.collectorWaitGroup.Wait()

	return nil
}

// Increase implements provider.InstanceGroup.
func (ig *InstanceGroup) Increase(ctx context.Context, count int) (int, error) {
	template, err := ig.getProxmoxVM(ctx, *ig.Settings.TemplateID)
	if err != nil {
		return 0, fmt.Errorf("failed to find template with id='%d': %w", *ig.Settings.TemplateID, err)
	}

	var (
		errorGroup = new(errgroup.Group)

		succeeded   = 0
		succeededMu = new(sync.Mutex)

		// We need to mutex clonning as Proxmox will fail multiple requests in parallel
		cloneMu = new(sync.Mutex)
	)

	ig.instanceClonningMu.Lock()
	defer ig.instanceClonningMu.Unlock()

	for n := 0; n < count; n++ {
		errorGroup.Go(func() error {
			vmid, err := ig.deployInstance(ctx, template, cloneMu)
			if err != nil {
				ig.log.Error("failed to deploy an instance", "vmid", vmid, "err", err)
			}

			ig.log.Info("successfully deployed instance", "vmid", vmid)
			succeededMu.Lock()
			succeeded++
			succeededMu.Unlock()

			return err
		})
	}

	if err := errorGroup.Wait(); err != nil {
		return succeeded, fmt.Errorf("failed to create one or more instances: %w", err)
	}

	return succeeded, nil
}

// Update implements provider.InstanceGroup.
func (ig *InstanceGroup) Update(ctx context.Context, update func(instance string, state provider.State)) error {
	ig.instanceClonningMu.Lock()
	defer ig.instanceClonningMu.Unlock()

	pool, err := ig.getProxmoxPool(ctx)
	if err != nil {
		return err
	}

	for _, member := range pool.Members {
		if !ig.isProxmoxResourceAnInstance(member) {
			continue
		}

		var state provider.State

		switch member.Name {
		case ig.Settings.InstanceNameCreating:
			state = provider.StateCreating
		case ig.Settings.InstanceNameRunning:
			state = provider.StateRunning
		case ig.Settings.InstanceNameRemoving:
			state = provider.StateDeleting
		default:
			continue // Unknown name, skipping...
		}

		update(strconv.FormatUint(member.VMID, 10), state)
	}

	return nil
}

// ConnectInfo implements provider.InstanceGroup.
func (ig *InstanceGroup) ConnectInfo(ctx context.Context, instance string) (provider.ConnectInfo, error) {
	VMID, err := strconv.Atoi(instance)
	if err != nil {
		return provider.ConnectInfo{}, fmt.Errorf("failed to parse instance name '%s': %w", instance, err)
	}

	vm, err := ig.getProxmoxVM(ctx, VMID)
	if err != nil {
		return provider.ConnectInfo{}, fmt.Errorf("failed to retrieve instance vmid='%d': %w", VMID, err)
	}

	networkInterfaces, err := vm.AgentGetNetworkIFaces(ctx)
	if err != nil {
		return provider.ConnectInfo{}, fmt.Errorf("failed to retrieve instance vmid='%d' interfaces: %w", VMID, err)
	}

	for _, networkInterface := range networkInterfaces {
		if networkInterface.Name != ig.Settings.InstanceNetworkInterface {
			continue
		}

		for _, address := range networkInterface.IPAddresses {
			if address.IPAddressType != "ipv4" {
				continue
			}

			return provider.ConnectInfo{
				ID:              instance,
				InternalAddr:    address.IPAddress,
				ExternalAddr:    address.IPAddress,
				ConnectorConfig: ig.FleetingSettings.ConnectorConfig,
			}, nil
		}
	}

	return provider.ConnectInfo{}, ErrNoIPv4Address
}

// Decrease implements provider.InstanceGroup.
func (ig *InstanceGroup) Decrease(ctx context.Context, instancesToRemove []string) ([]string, error) {
	pool, err := ig.getProxmoxPool(ctx)
	if err != nil {
		return []string{}, err
	}

	var (
		errorGroup = new(errgroup.Group)

		succeeded   = []string{}
		succeededMu = new(sync.Mutex)
	)

	for _, member := range pool.Members {
		member := member

		if !ig.isProxmoxResourceAnInstance(member) {
			continue
		}

		if !slices.Contains(instancesToRemove, strconv.FormatUint(member.VMID, 10)) {
			continue
		}

		if member.Name == ig.Settings.InstanceNameCreating {
			// It must be running to start the deletion
			continue
		}

		if member.Name == ig.Settings.InstanceNameRemoving {
			// Already deleting...
			succeededMu.Lock()
			succeeded = append(succeeded, strconv.FormatUint(member.VMID, 10))
			succeededMu.Unlock()

			continue
		}

		ig.log.Info("removing instance", "vmid", member.VMID)

		errorGroup.Go(func() error {
			if err := ig.markInstancesForRemoval(ctx, &member); err != nil {
				return err
			}

			succeededMu.Lock()
			defer succeededMu.Unlock()
			succeeded = append(succeeded, strconv.FormatUint(member.VMID, 10))

			return nil
		})
	}

	//nolint:wrapcheck
	return succeeded, errorGroup.Wait()
}
