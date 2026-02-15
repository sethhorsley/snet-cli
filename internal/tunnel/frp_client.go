package tunnel

import (
	"context"
	"fmt"
	"time"

	"github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/seth4242/snet/internal/api"
)

// EmbeddedFRPClient wraps the FRP client library
type EmbeddedFRPClient struct {
	cfg    *v1.ClientConfig
	svr    *client.Service
	tunnel *api.Tunnel
	port   int
	host   string
}

// NewEmbeddedFRPClient creates a new embedded FRP client
func NewEmbeddedFRPClient(tunnel *api.Tunnel, port int, host string) (*EmbeddedFRPClient, error) {
	if host == "" {
		host = "127.0.0.1"
	}

	// Create HTTP proxy configuration
	httpProxy := &v1.HTTPProxyConfig{
		ProxyBaseConfig: v1.ProxyBaseConfig{
			Name: tunnel.FRPProxyName,
			Type: "http",
			ProxyBackend: v1.ProxyBackend{
				LocalIP:   host,
				LocalPort: port,
			},
		},
		DomainConfig: v1.DomainConfig{
			SubDomain: tunnel.Slug,
		},
	}

	// Configure FRP client
	cfg := &v1.ClientConfig{
		ClientCommonConfig: v1.ClientCommonConfig{
			ServerAddr: "149.248.211.110",
			ServerPort: 7000,
			Auth: v1.AuthClientConfig{
				Method: v1.AuthMethodToken,
				Token:  tunnel.FRPAuthToken,
			},
			Log: v1.LogConfig{
				Level: "info",
				To:    "console",
			},
		},
		Proxies: []v1.TypedProxyConfig{
			{
				Type:            "http",
				ProxyConfigurer: httpProxy,
			},
		},
	}

	return &EmbeddedFRPClient{
		cfg:    cfg,
		tunnel: tunnel,
		port:   port,
		host:   host,
	}, nil
}

// Start starts the embedded FRP client
func (c *EmbeddedFRPClient) Start(ctx context.Context) error {
	// Complete the configuration (sets defaults)
	c.cfg.Complete()

	// Convert TypedProxyConfig to ProxyConfigurer
	proxyCfgs := make([]v1.ProxyConfigurer, 0, len(c.cfg.Proxies))
	for i := range c.cfg.Proxies {
		proxyCfgs = append(proxyCfgs, c.cfg.Proxies[i].ProxyConfigurer)
	}

	// Convert TypedVisitorConfig to VisitorConfigurer
	visitorCfgs := make([]v1.VisitorConfigurer, 0, len(c.cfg.Visitors))
	for i := range c.cfg.Visitors {
		visitorCfgs = append(visitorCfgs, c.cfg.Visitors[i].VisitorConfigurer)
	}

	var err error
	c.svr, err = client.NewService(client.ServiceOptions{
		Common:      &c.cfg.ClientCommonConfig,
		ProxyCfgs:   proxyCfgs,
		VisitorCfgs: visitorCfgs,
	})
	if err != nil {
		return fmt.Errorf("failed to create FRP service: %w", err)
	}

	// Run the service
	if err := c.svr.Run(ctx); err != nil {
		return fmt.Errorf("FRP client error: %w", err)
	}

	return nil
}

// Stop stops the embedded FRP client
func (c *EmbeddedFRPClient) Stop() {
	if c.svr != nil {
		c.svr.Close()
	}
}

// Wait waits for the FRP client to exit
func (c *EmbeddedFRPClient) Wait() error {
	// The FRP service runs until context is cancelled
	// Just wait a bit to ensure it's running
	time.Sleep(100 * time.Millisecond)
	return nil
}

// IsConnected checks if the FRP client is connected
func (c *EmbeddedFRPClient) IsConnected() bool {
	// TODO: Check actual connection status from FRP service
	return c.svr != nil
}
