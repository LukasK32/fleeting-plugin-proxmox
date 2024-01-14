package plugin

import (
	"context"
	"errors"

	hclog "github.com/hashicorp/go-hclog"
	"gitlab.com/gitlab-org/fleeting/fleeting/provider"
)

var _ provider.InstanceGroup = (*InstanceGroup)(nil)

type InstanceGroup struct{}

// Init implements provider.InstanceGroup.
func (ig *InstanceGroup) Init(ctx context.Context, logger hclog.Logger, settings provider.Settings) (provider.ProviderInfo, error) {
	// FIXME: Implement
	return provider.ProviderInfo{}, errors.New("method Init is unimplemented")
}

// Shutdown implements provider.InstanceGroup.
func (ig *InstanceGroup) Shutdown(ctx context.Context) error {
	// TODO: Verify if we need to do anything on shutdown
	return nil
}

// Increase implements provider.InstanceGroup.
func (ig *InstanceGroup) Increase(ctx context.Context, n int) (succeeded int, err error) {
	// FIXME: Implement
	return 0, errors.New("method Increase is unimplemented")
}

// Update implements provider.InstanceGroup.
func (ig *InstanceGroup) Update(ctx context.Context, fn func(instance string, state provider.State)) error {
	// FIXME: Implement
	return errors.New("method Update is unimplemented")
}

// ConnectInfo implements provider.InstanceGroup.
func (ig *InstanceGroup) ConnectInfo(ctx context.Context, instance string) (provider.ConnectInfo, error) {
	// FIXME: Implement
	return provider.ConnectInfo{}, errors.New("method ConnectInfo is unimplemented")
}

// Decrease implements provider.InstanceGroup.
func (ig *InstanceGroup) Decrease(ctx context.Context, instances []string) (succeeded []string, err error) {
	// FIXME: Implement
	return []string{}, errors.New("method Decrease is unimplemented")
}
