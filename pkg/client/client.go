package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/operator-framework/operator-registry/pkg/api"
	"github.com/operator-framework/operator-registry/pkg/api/grpc_health_v1"
	"github.com/operator-framework/operator-registry/pkg/registry"
)

type Interface interface {
	GetBundle(ctx context.Context, packageName, channelName, csvName string) (*registry.Bundle, error)
	GetBundlePath(ctx context.Context, packageName, channelName, csvName string) (string, error)
	GetBundleInPackageChannel(ctx context.Context, packageName, channelName string) (*registry.Bundle, error)
	GetReplacementBundleInPackageChannel(ctx context.Context, currentName, packageName, channelName string) (*registry.Bundle, error)
	GetBundleThatProvides(ctx context.Context, group, version, kind string) (*registry.Bundle, error)
	HealthCheck(ctx context.Context, reconnectTimeout time.Duration) (bool, error)
	Close() error
}

type Client struct {
	Registry api.RegistryClient
	Health   grpc_health_v1.HealthClient
	Conn     *grpc.ClientConn
}

var _ Interface = &Client{}

func (c *Client) GetBundle(ctx context.Context, packageName, channelName, csvName string) (*registry.Bundle, error) {
	bundle, err := c.Registry.GetBundle(ctx, &api.GetBundleRequest{PkgName: packageName, ChannelName: channelName, CsvName: csvName})
	if err != nil {
		return nil, err
	}
	return registry.NewBundleFromStrings(bundle.CsvName, bundle.PackageName, bundle.ChannelName, bundle.Object)
}

func (c *Client) GetBundlePath(ctx context.Context, packageName, channelName, csvName string) (string, error) {
	bundlePath, err := c.Registry.GetBundlePath(ctx, &api.GetBundlePathRequest{PkgName: packageName, ChannelName: channelName, CsvName: csvName})
	if err != nil {
		return "", err
	}
	return bundlePath.Path, nil
}

func (c *Client) GetBundleInPackageChannel(ctx context.Context, packageName, channelName string) (*registry.Bundle, error) {
	bundle, err := c.Registry.GetBundleForChannel(ctx, &api.GetBundleInChannelRequest{PkgName: packageName, ChannelName: channelName})
	if err != nil {
		return nil, err
	}
	return registry.NewBundleFromStrings(bundle.CsvName, packageName, channelName, bundle.Object)
}

func (c *Client) GetReplacementBundleInPackageChannel(ctx context.Context, currentName, packageName, channelName string) (*registry.Bundle, error) {
	bundle, err := c.Registry.GetBundleThatReplaces(ctx, &api.GetReplacementRequest{CsvName: currentName, PkgName: packageName, ChannelName: channelName})
	if err != nil {
		return nil, err
	}
	return registry.NewBundleFromStrings(bundle.CsvName, packageName, channelName, bundle.Object)
}

func (c *Client) GetBundleThatProvides(ctx context.Context, group, version, kind string) (*registry.Bundle, error) {
	bundle, err := c.Registry.GetDefaultBundleThatProvides(ctx, &api.GetDefaultProviderRequest{Group: group, Version: version, Kind: kind})
	if err != nil {
		return nil, err
	}
	parsedBundle, err := registry.NewBundleFromStrings(bundle.CsvName, bundle.PackageName, bundle.ChannelName, bundle.Object)
	if err != nil {
		return nil, err
	}
	return parsedBundle, nil
}

func (c *Client) Close() error {
	if c.Conn == nil {
		return nil
	}
	return c.Conn.Close()
}

func (c *Client) HealthCheck(ctx context.Context, reconnectTimeout time.Duration) (bool, error) {
	res, err := c.Health.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "Registry"})
	if err != nil {
		if c.Conn.GetState() == connectivity.TransientFailure {
			ctx, cancel := context.WithTimeout(ctx, reconnectTimeout)
			defer cancel()
			if !c.Conn.WaitForStateChange(ctx, connectivity.TransientFailure) {
				return false, NewHealthError(c.Conn, HealthErrReasonUnrecoveredTransient, "connection didn't recover from TransientFailure")
			}
		}
		return false, NewHealthError(c.Conn, HealthErrReasonConnection, err.Error())
	}
	if res.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return false, nil
	}
	return true, nil
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return NewClientFromConn(conn), nil
}

func NewClientFromConn(conn *grpc.ClientConn) *Client {
	return &Client{
		Registry: api.NewRegistryClient(conn),
		Health:   grpc_health_v1.NewHealthClient(conn),
		Conn:     conn,
	}
}
