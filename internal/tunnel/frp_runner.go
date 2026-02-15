package tunnel

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/seth4242/snet/internal/api"
)

// FRPRunner manages the embedded FRP client and heartbeats
type FRPRunner struct {
	client       *api.Client
	tunnel       *api.Tunnel
	port         int
	host         string
	connectionID int
	frpClient    *EmbeddedFRPClient
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewFRPRunner creates a new FRP tunnel runner
func NewFRPRunner(client *api.Client, tunnel *api.Tunnel, port int, host string) *FRPRunner {
	ctx, cancel := context.WithCancel(context.Background())
	if host == "" {
		host = "127.0.0.1"
	}
	return &FRPRunner{
		client: client,
		tunnel: tunnel,
		port:   port,
		host:   host,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run starts the FRP tunnel and blocks until it exits
func (r *FRPRunner) Run() error {
	// Create embedded FRP client
	frpClient, err := NewEmbeddedFRPClient(r.tunnel, r.port, r.host)
	if err != nil {
		return fmt.Errorf("failed to create FRP client: %w", err)
	}
	r.frpClient = frpClient

	fmt.Println("\nStarting embedded FRP client...")

	// Report connection to API
	if err := r.reportConnect(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to report connection: %v\n", err)
	}

	// Start heartbeat goroutine
	r.wg.Add(1)
	go r.heartbeatLoop()

	// Start FRP client in a goroutine
	r.wg.Add(1)
	errChan := make(chan error, 1)
	go func() {
		defer r.wg.Done()
		if err := r.frpClient.Start(r.ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Wait for FRP client to be ready
	time.Sleep(1 * time.Second)

	if r.frpClient.IsConnected() {
		fmt.Println("✓ FRP client connected successfully!")
	} else {
		fmt.Println("⚠ FRP client starting...")
	}
	fmt.Println("")

	// Show tunnel URLs with SSL status
	r.showTunnelStatus()

	fmt.Println("\nProxy running. Press Ctrl+C to stop...")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
	case err := <-errChan:
		fmt.Fprintf(os.Stderr, "\nFRP client error: %v\n", err)
		r.shutdown()
		r.wg.Wait()
		return err
	}

	r.shutdown()
	r.wg.Wait()
	return nil
}

// reportConnect reports the connection to the API
func (r *FRPRunner) reportConnect() error {
	resp, err := r.client.Connect(r.tunnel.ID, &api.ConnectRequest{
		CloudflaredVersion: fmt.Sprintf("snet-embedded-frp-v0.61.1-%s-%s", runtime.GOOS, runtime.GOARCH),
		Metadata: map[string]interface{}{
			"provider": "frp",
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"port":     r.port,
			"host":     r.host,
			"mode":     "embedded", // Embedded FRP client
		},
	})
	if err != nil {
		return err
	}
	r.connectionID = resp.ConnectionID
	return nil
}

// reportDisconnect reports the disconnection to the API
func (r *FRPRunner) reportDisconnect() {
	if r.connectionID == 0 {
		return
	}

	err := r.client.Disconnect(r.tunnel.ID, &api.DisconnectRequest{
		ConnectionID: r.connectionID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to report disconnection: %v\n", err)
	}
}

// heartbeatLoop sends heartbeats every 30 seconds
func (r *FRPRunner) heartbeatLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := r.client.Heartbeat(r.tunnel.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: heartbeat failed: %v\n", err)
			}
		case <-r.ctx.Done():
			return
		}
	}
}

// showTunnelStatus displays tunnel URLs with SSL validation status
func (r *FRPRunner) showTunnelStatus() {
	// Main URL is always ready immediately (DNS points to FRP server)
	fmt.Printf("✓ Main URL:     %s\n", r.tunnel.URL)

	// If wildcard enabled, check SSL status
	if r.tunnel.Wildcard && r.tunnel.WildcardURL != "" {
		// Check SSL status
		status, err := r.client.SSLStatus(r.tunnel.ID)
		if err != nil {
			fmt.Printf("  Wildcard URL: %s (checking SSL...)\n", r.tunnel.WildcardURL)
		} else if status.WildcardReady {
			fmt.Printf("✓ Wildcard URL: %s\n", r.tunnel.WildcardURL)
		} else {
			fmt.Printf("⏳ Wildcard URL: %s (waiting for SSL validation...)\n", r.tunnel.WildcardURL)
			// Start background SSL status checker
			r.wg.Add(1)
			go r.pollSSLStatus()
		}
	}
}

// pollSSLStatus polls for SSL validation in the background
func (r *FRPRunner) pollSSLStatus() {
	defer r.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := r.client.SSLStatus(r.tunnel.ID)
			if err != nil {
				// Silently continue polling
				continue
			}

			if status.WildcardReady {
				fmt.Printf("\n✓ Wildcard SSL validated! %s is now ready\n", r.tunnel.WildcardURL)
				return
			}
		case <-r.ctx.Done():
			return
		}
	}
}

// shutdown gracefully stops the tunnel
func (r *FRPRunner) shutdown() {
	// Stop FRP client
	if r.frpClient != nil {
		r.frpClient.Stop()
	}

	// Cancel context to stop heartbeat
	r.cancel()

	// Report disconnection
	r.reportDisconnect()

	// Delete tunnel if it's ephemeral (not persistent)
	// This triggers immediate cleanup of DNS records and SSL certificates
	if !r.tunnel.Persistent {
		fmt.Println("Cleaning up tunnel resources...")
		err := r.client.DeleteTunnel(r.tunnel.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete tunnel: %v\n", err)
		} else {
			fmt.Println("✓ Tunnel resources cleaned up")
		}
	}
}
