package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/luthermonson/go-proxmox"
	"golang.org/x/sync/errgroup"
)

const (
	proxmoxTaskWaitInterval  = 10 * time.Second
	proxmoxTaskWaitTimeout   = 5 * time.Minute
	proxmoxAgentStartTimeout = 2 * time.Minute
)

func (ig *InstanceGroup) deployInstance(ctx context.Context, template *proxmox.VirtualMachine, cloneMu *sync.Mutex) (int, error) {
	cloneMu.Lock()
	VMID, task, err := template.Clone(
		ctx,
		&proxmox.VirtualMachineCloneOptions{
			Name:    ig.PluginSettings.InstanceNameCreating,
			Pool:    ig.PluginSettings.Pool,
			Storage: ig.PluginSettings.Storage,
		},
	)
	cloneMu.Unlock()

	if err == nil {
		ig.log.Info("Deploying new instance", "vmid", VMID)

		err = task.Wait(ctx, proxmoxTaskWaitInterval, proxmoxTaskWaitTimeout)
	}

	if err != nil {
		return VMID, fmt.Errorf("failed to clone template: %w", err)
	}

	vm, err := ig.getProxmoxVM(ctx, VMID)
	if err != nil {
		return VMID, fmt.Errorf("failed to find newly deployed instance vmid='%d': %w", VMID, err)
	}

	// Start, configure etc.
	err = func() error {
		// Start the VM
		task, err := vm.Start(ctx)
		if err == nil {
			err = task.Wait(ctx, proxmoxTaskWaitInterval, proxmoxTaskWaitTimeout)
		}

		if err != nil {
			return fmt.Errorf("failed to start instance: %w", err)
		}

		// Wait for agent to start
		if err := vm.WaitForAgent(ctx, int(proxmoxAgentStartTimeout/time.Second)); err != nil {
			return fmt.Errorf("failed when waiting for qemu agent to start: %w", err)
		}

		return nil
	}()

	newInstanceName := ig.PluginSettings.InstanceNameRunning

	if err != nil {
		ig.log.Error("instance deployment failed, marking for removal", "vmid", VMID, "err", err)
		newInstanceName = ig.PluginSettings.InstanceNameRemoving
	}

	_, renameErr := vm.Config(ctx, proxmox.VirtualMachineOption{
		Name:  "name",
		Value: newInstanceName,
	})

	if renameErr != nil {
		ig.log.Error("failed to rename instance", "vmid", VMID, "err", renameErr)
	}

	if err != nil {
		return VMID, fmt.Errorf("failed to configure instance, marked for removal due to: %w", err)
	}

	return VMID, nil
}

func (ig *InstanceGroup) markStaleInstancesForRemoval(ctx context.Context) error {
	pool, err := ig.getProxmoxPool(ctx)
	if err != nil {
		return err
	}

	instancesToMarkForRemoval := make([]*proxmox.ClusterResource, 0, len(pool.Members))

	for _, member := range pool.Members {
		member := member

		if !ig.isProxmoxResourceAnInstance(member) {
			continue
		}

		if member.Name != ig.PluginSettings.InstanceNameCreating {
			continue
		}

		ig.log.Info("Found stale instance, marking for removal", "name", member.Name, "vmid", member.VMID, "node", member.Node)
		instancesToMarkForRemoval = append(instancesToMarkForRemoval, &member)
	}

	if len(instancesToMarkForRemoval) < 1 {
		return nil
	}

	if err := ig.markInstancesForRemoval(ctx, instancesToMarkForRemoval...); err != nil {
		return fmt.Errorf("failed to mark stale instances for removal: %w", err)
	}

	return nil
}

func (ig *InstanceGroup) markInstancesForRemoval(ctx context.Context, instances ...*proxmox.ClusterResource) error {
	var errorGroup errgroup.Group

	for _, instance := range instances {
		instance := instance

		errorGroup.Go(func() error {
			log := ig.log.With("name", instance.Name, "vmid", instance.VMID, "node", instance.Node)

			vm, err := ig.getProxmoxVMOnNode(ctx, int(instance.VMID), instance.Node)
			if err != nil {
				log.Error("Failed to mark instance for removal", "err", err)
				return fmt.Errorf("failed to mark instance for removal: %w", err)
			}

			task, err := vm.Config(ctx, proxmox.VirtualMachineOption{
				Name:  "name",
				Value: ig.PluginSettings.InstanceNameRemoving,
			})

			if err == nil {
				err = task.Wait(ctx, proxmoxTaskWaitInterval, proxmoxTaskWaitTimeout)
			}

			if err != nil {
				log.Error("Failed to mark instance for removal", "err", err)
				return fmt.Errorf("failed to mark instance for removal: %w", err)
			}

			return nil
		})
	}

	if err := errorGroup.Wait(); err != nil {
		ig.instanceCollectionTrigger <- struct{}{}
		return fmt.Errorf("failed to mark one or more instances for removal: %w", err)
	}

	ig.instanceCollectionTrigger <- struct{}{}

	return nil
}

func (ig *InstanceGroup) isProxmoxResourceAnInstance(member proxmox.ClusterResource) bool {
	return member.Type == "qemu" && member.VMID != uint64(*ig.PluginSettings.TemplateID)
}
