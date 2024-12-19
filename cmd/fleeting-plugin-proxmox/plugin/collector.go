package plugin

import (
	"context"
	"sync"
	"time"

	"github.com/luthermonson/go-proxmox"
)

const (
	collectionInterval         = 1 * time.Minute
	collectionTimeout          = 5 * time.Minute
	collectionWaitAfterTrigger = 10 * time.Second
)

func (ig *InstanceGroup) startRemovedInstanceCollector() {
	ig.collectorWaitGroup.Add(1)

	go func() {
		defer ig.collectorWaitGroup.Done()
		ig.runRemovedInstanceCollector()
	}()
}

func (ig *InstanceGroup) runRemovedInstanceCollector() {
	ig.collectRemovedInstances()

	for {
		select {
		case <-ig.collectorShutdownTrigger:
			return
		case <-time.After(collectionInterval):
			ig.collectRemovedInstances()
		case <-ig.instanceCollectionTrigger:
			ig.drainInstanceCollectionTriggerChannel()

			// Sleep for a bit to give Proxmox a chance to propagate renames that happened before trigger
			<-time.After(collectionWaitAfterTrigger)

			ig.collectRemovedInstances()
		}
	}
}

func (ig *InstanceGroup) collectRemovedInstances() {
	ctx, cancel := context.WithTimeout(context.Background(), collectionTimeout)
	defer cancel()

	ig.instanceCloningMu.Lock()

	pool, err := ig.getProxmoxPool(ctx)
	if err != nil {
		ig.log.Error("collector failed to list instances", "err", err)
		ig.instanceCloningMu.Unlock()

		return
	}

	ig.instanceCloningMu.Unlock()

	var wg sync.WaitGroup
	defer wg.Wait()

	for _, member := range pool.Members {
		if !ig.isProxmoxResourceAnInstance(member) {
			continue
		}

		if member.Name != ig.Settings.InstanceNameRemoving {
			continue
		}

		ig.log.Info("collector found instance to remove", "vmid", member.VMID, "name", member.Name)

		wg.Add(1)

		go func(member proxmox.ClusterResource) {
			defer wg.Done()
			ig.collectInstance(ctx, member)
		}(member)
	}
}

func (ig *InstanceGroup) collectInstance(ctx context.Context, member proxmox.ClusterResource) {
	vm, err := ig.getProxmoxVMOnNode(ctx, int(member.VMID), member.Node)
	if err != nil {
		ig.log.Error("collector failed to fetch instance info", "vmid", member.VMID, "err", err)
		return
	}

	if vm.Status == "running" {
		task, err := vm.Stop(ctx)
		if err == nil {
			err = task.Wait(ctx, proxmoxTaskWaitInterval, collectionTimeout)
		}

		if err != nil {
			ig.log.Error("collector failed to stop instance", "vmid", member.VMID, "err", err)
			return
		}
	}

	task, err := vm.Delete(ctx)
	if err == nil {
		err = task.Wait(ctx, proxmoxTaskWaitInterval, collectionTimeout)
	}

	if err != nil {
		ig.log.Error("collector failed to delete instance", "vmid", member.VMID, "err", err)
	}
}

func (ig *InstanceGroup) drainInstanceCollectionTriggerChannel() {
	for {
		select {
		case <-ig.instanceCollectionTrigger:
			// NOOP
		default:
			return
		}
	}
}
