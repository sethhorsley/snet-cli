package tunnel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/seth4242/snet/internal/api"
)

// Runner manages the cloudflared process and heartbeats
type Runner struct {
	client       *api.Client
	tunnel       *api.Tunnel
	port         int
	host         string
	connectionID int
	cmd          *exec.Cmd
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewRunner creates a new tunnel runner
func NewRunner(client *api.Client, tunnel *api.Tunnel, port int, host string) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	if host == "" {
		host = "localhost"
	}
	return &Runner{
		client: client,
		tunnel: tunnel,
		port:   port,
		host:   host,
		ctx:    ctx,
		cancel: cancel,
	}
}

// CheckCloudflared verifies cloudflared is installed
func CheckCloudflared() error {
	_, err := exec.LookPath("cloudflared")
	if err != nil {
		return fmt.Errorf(`cloudflared not found in PATH

Install cloudflared:
  brew install cloudflared

Or download from:
  https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/`)
	}
	return nil
}

// GetCloudflaredVersion returns the installed cloudflared version
func GetCloudflaredVersion() string {
	cmd := exec.Command("cloudflared", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return string(output)
}

// Run starts the tunnel and blocks until it exits
func (r *Runner) Run() error {
	// Check cloudflared is installed
	if err := CheckCloudflared(); err != nil {
		return err
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start cloudflared
	if err := r.startCloudflared(); err != nil {
		return fmt.Errorf("failed to start cloudflared: %w", err)
	}

	// Report connection to API
	if err := r.reportConnect(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to report connection: %v\n", err)
	}

	// Start heartbeat goroutine
	r.wg.Add(1)
	go r.heartbeatLoop()

	// Wait for signal or cloudflared to exit
	cloudflaredDone := make(chan error, 1)
	go func() {
		cloudflaredDone <- r.cmd.Wait()
	}()

	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
		r.shutdown()
	case err := <-cloudflaredDone:
		fmt.Println("\ncloudflared exited")
		r.cancel()
		r.reportDisconnect()
		if err != nil {
			return fmt.Errorf("cloudflared exited with error: %w", err)
		}
	}

	r.wg.Wait()
	return nil
}

// startCloudflared starts the cloudflared process
func (r *Runner) startCloudflared() error {
	r.cmd = exec.Command("cloudflared", "tunnel", "--no-autoupdate", "run",
		"--token", r.tunnel.CloudflareTunnelToken,
		"--url", fmt.Sprintf("http://%s:%d", r.host, r.port))

	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr

	return r.cmd.Start()
}

// reportConnect reports the connection to the API
func (r *Runner) reportConnect() error {
	resp, err := r.client.Connect(r.tunnel.ID, &api.ConnectRequest{
		CloudflaredVersion: GetCloudflaredVersion(),
		Metadata: map[string]interface{}{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
	})
	if err != nil {
		return err
	}
	r.connectionID = resp.ConnectionID
	return nil
}

// reportDisconnect reports the disconnection to the API
func (r *Runner) reportDisconnect() {
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
func (r *Runner) heartbeatLoop() {
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

// shutdown gracefully stops the tunnel
func (r *Runner) shutdown() {
	// Cancel context to stop heartbeat
	r.cancel()

	// Stop cloudflared
	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Signal(syscall.SIGTERM)

		// Give it a moment to exit gracefully
		done := make(chan struct{})
		go func() {
			r.cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			r.cmd.Process.Kill()
		}
	}

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
