package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/util/log"
	"github.com/seth4242/snet/internal/api"
)

// TUILogger is a callback interface for sending FRP logs to the TUI
type TUILogger interface {
	LogHTTPRequest(method, path string, statusCode int, duration time.Duration, isAppLog bool)
}

// EmbeddedFRPClient wraps the FRP client library
type EmbeddedFRPClient struct {
	cfg         *v1.ClientConfig
	svr         *client.Service
	tunnel      *api.Tunnel
	port        int
	host        string
	suppressLog bool
	tuiLogger   TUILogger
	logFile     string // Path to temp log file for capturing FRP logs
}

// NewEmbeddedFRPClient creates a new embedded FRP client
func NewEmbeddedFRPClient(tunnel *api.Tunnel, port int, host string, suppressLog bool, tuiLogger TUILogger, headerConfig *HeaderConfig) (*EmbeddedFRPClient, error) {
	if host == "" {
		host = "127.0.0.1"
	}

	// Create HTTP proxy configuration
	// Use customDomains to specify full domain name: {slug}.{account}.seth4242.net
	// FRP server has no subDomainHost configured, allowing customDomains to work
	customDomains := []string{
		fmt.Sprintf("%s.%s.seth4242.net", tunnel.Slug, tunnel.Account.Slug),
	}

	// Add wildcard domain if enabled
	if tunnel.Wildcard {
		customDomains = append(customDomains, fmt.Sprintf("*.%s.%s.seth4242.net", tunnel.Slug, tunnel.Account.Slug))
	}

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
			CustomDomains: customDomains,
		},
		// Transparent mode by default - no header manipulation
	}

	// Apply custom header configuration if provided
	headerConfig.ApplyToProxyConfig(httpProxy)

	// Determine log configuration
	// In TUI mode: write logs to a temp file so we can tail and display them when 't' is toggled
	// In plain mode: show info logs to console
	logLevel := "info"
	logDestination := "console"
	var logFile string

	if suppressLog && tuiLogger != nil {
		// In TUI mode: write logs to a temp file that we'll tail and send to TUI
		tmpFile, err := os.CreateTemp("", "snet-frp-*.log")
		if err != nil {
			// Fallback to /dev/null if temp file creation fails
			logDestination = "/dev/null"
		} else {
			logDestination = tmpFile.Name()
			logFile = tmpFile.Name()
			tmpFile.Close() // FRP will open it for writing
		}
		// Keep info level to capture all FRP connection logs
		logLevel = "info"
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
				Level: logLevel,
				To:    logDestination,
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
		cfg:         cfg,
		tunnel:      tunnel,
		port:        port,
		host:        host,
		suppressLog: suppressLog,
		tuiLogger:   tuiLogger,
		logFile:     logFile,
	}, nil
}

// Start starts the embedded FRP client
func (c *EmbeddedFRPClient) Start(ctx context.Context) error {
	// Complete the configuration (sets defaults)
	c.cfg.Complete()

	// CRITICAL: Initialize the global FRP logger
	// The FRP library uses a global logger that must be explicitly initialized.
	// When embedding the library, this doesn't happen automatically like it does in the CLI.
	// Without this, FRP will use the default logger which writes to stdout at info level.
	log.InitLogger(
		c.cfg.Log.To,
		c.cfg.Log.Level,
		int(c.cfg.Log.MaxDays),
		c.cfg.Log.DisablePrintColor,
	)

	// If we have a log file and TUI logger, start tailing the log file
	if c.logFile != "" && c.tuiLogger != nil {
		go c.tailLogFile(ctx)
	}

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

// tailLogFile tails the FRP log file and sends parsed logs to the TUI
func (c *EmbeddedFRPClient) tailLogFile(ctx context.Context) {
	// Wait a moment for the log file to be created and written
	time.Sleep(500 * time.Millisecond)

	file, err := os.Open(c.logFile)
	if err != nil {
		return
	}
	defer file.Close()

	// Start reading from the beginning to capture all logs (including initial connection logs)
	// Don't seek to end - we want to capture the connection messages that were already written

	scanner := bufio.NewScanner(file)
	// FRP log format: 2026-02-16 10:37:37.186 [I] [client/service.go:295] try to connect to server...
	logPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \[[A-Z]\] \[([^\]]+)\] (.+)$`)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if scanner.Scan() {
				line := scanner.Text()

				// Parse the log line
				matches := logPattern.FindStringSubmatch(line)
				if len(matches) >= 3 {
					// matches[1] is the source (e.g., "client/service.go:295")
					// matches[2] is the message
					message := matches[2]

					// Send to TUI as infrastructure log (IsAppLog = false)
					c.tuiLogger.LogHTTPRequest("FRP", message, 0, 0, false)
				} else if line != "" {
					// Didn't match pattern, send as-is if not empty
					c.tuiLogger.LogHTTPRequest("FRP", line, 0, 0, false)
				}
			} else {
				// No more lines, wait before checking again
				time.Sleep(100 * time.Millisecond)
				if err := scanner.Err(); err != nil {
					return
				}
			}
		}
	}
}
