package plugin

import (
	"context"
	"time"
)

const (
	sessionTicketRefreshInterval = 1 * time.Hour
	sessionTicketRefreshTimeout  = 5 * time.Second
)

func (ig *InstanceGroup) startSessionTicketRefresher() {
	ig.sessionTicketRefresherWaitGroup.Add(1)

	go func() {
		defer ig.sessionTicketRefresherWaitGroup.Done()
		ig.runSessionTicketRefresher()
	}()
}

func (ig *InstanceGroup) runSessionTicketRefresher() {
	for {
		select {
		case <-ig.sessionTicketRefresherShutdownTrigger:
			return
		case <-time.After(sessionTicketRefreshInterval):
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), sessionTicketRefreshTimeout)
				defer cancel()

				credentials, err := ig.getProxmoxCredentials()
				if err != nil {
					ig.log.Error("failed to refresh proxmox session, could not read credentials", "err", err)
					return
				}

				_, err = ig.proxmox.Ticket(ctx, credentials)
				if err != nil {
					ig.log.Error("failed to refresh proxmox session", "err", err)
				}

				ig.log.Info("refreshed proxmox session")
			}()
		}
	}
}
