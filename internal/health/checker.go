package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Checker performs health checks on services
type Checker struct {
	log      *logrus.Logger
	checks   map[string]*Check
	mu       sync.RWMutex
	stopChan chan struct{}
	running  bool
}

// Check represents a health check
type Check struct {
	ID       string
	Type     string // http, tcp, grpc
	Target   string
	Interval time.Duration
	Timeout  time.Duration
	Status   CheckStatus
	LastCheck time.Time
	Failures int
	callback func(status CheckStatus)
}

// CheckStatus represents the status of a health check
type CheckStatus string

const (
	StatusPassing  CheckStatus = "passing"
	StatusWarning  CheckStatus = "warning"
	StatusCritical CheckStatus = "critical"
)

// NewChecker creates a new health checker
func NewChecker(log *logrus.Logger) *Checker {
	return &Checker{
		log:      log,
		checks:   make(map[string]*Check),
		stopChan: make(chan struct{}),
	}
}

// Start starts the health checker
func (c *Checker) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("health checker is already running")
	}
	c.running = true
	c.mu.Unlock()
	
	c.log.Info("Starting health checker...")
	
	// Start check loops for all registered checks
	c.mu.RLock()
	for _, check := range c.checks {
		go c.checkLoop(ctx, check)
	}
	c.mu.RUnlock()
	
	return nil
}

// Stop stops the health checker
func (c *Checker) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return fmt.Errorf("health checker is not running")
	}
	c.running = false
	c.mu.Unlock()
	
	close(c.stopChan)
	c.log.Info("Health checker stopped")
	
	return nil
}

// AddCheck adds a new health check
func (c *Checker) AddCheck(check *Check) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if check.ID == "" {
		check.ID = generateCheckID()
	}
	
	if check.Interval == 0 {
		check.Interval = 10 * time.Second
	}
	
	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}
	
	check.Status = StatusPassing
	check.LastCheck = time.Now()
	
	c.checks[check.ID] = check
	c.log.Infof("Added health check: %s (%s)", check.ID, check.Type)
	
	// Start check loop if checker is running
	if c.running {
		go c.checkLoop(context.Background(), check)
	}
	
	return nil
}

// RemoveCheck removes a health check
func (c *Checker) RemoveCheck(checkID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.checks[checkID]; !exists {
		return fmt.Errorf("check not found: %s", checkID)
	}
	
	delete(c.checks, checkID)
	c.log.Infof("Removed health check: %s", checkID)
	
	return nil
}

// GetCheck returns a health check by ID
func (c *Checker) GetCheck(checkID string) (*Check, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	check, exists := c.checks[checkID]
	if !exists {
		return nil, fmt.Errorf("check not found: %s", checkID)
	}
	
	return check, nil
}

// ListChecks returns all health checks
func (c *Checker) ListChecks() []*Check {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	checks := make([]*Check, 0, len(c.checks))
	for _, check := range c.checks {
		checks = append(checks, check)
	}
	
	return checks
}

// checkLoop runs the health check loop for a specific check
func (c *Checker) checkLoop(ctx context.Context, check *Check) {
	ticker := time.NewTicker(check.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.performCheck(check)
		}
	}
}

// performCheck performs a single health check
func (c *Checker) performCheck(check *Check) {
	c.mu.Lock()
	check.LastCheck = time.Now()
	c.mu.Unlock()
	
	var err error
	
	switch check.Type {
	case "http":
		err = c.checkHTTP(check)
	case "tcp":
		err = c.checkTCP(check)
	case "grpc":
		err = c.checkGRPC(check)
	default:
		c.log.Errorf("Unknown check type: %s", check.Type)
		return
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if err != nil {
		check.Failures++
		if check.Failures >= 3 {
			check.Status = StatusCritical
		} else {
			check.Status = StatusWarning
		}
		c.log.Warnf("Health check failed: %s - %v", check.ID, err)
	} else {
		check.Failures = 0
		check.Status = StatusPassing
		c.log.Debugf("Health check passed: %s", check.ID)
	}
	
	// Call callback if set
	if check.callback != nil {
		go check.callback(check.Status)
	}
}

// checkHTTP performs an HTTP health check
func (c *Checker) checkHTTP(check *Check) error {
	client := &http.Client{
		Timeout: check.Timeout,
	}
	
	resp, err := client.Get(check.Target)
	if err != nil {
		return fmt.Errorf("HTTP check failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP check returned status %d", resp.StatusCode)
	}
	
	return nil
}

// checkTCP performs a TCP health check
func (c *Checker) checkTCP(check *Check) error {
	conn, err := net.DialTimeout("tcp", check.Target, check.Timeout)
	if err != nil {
		return fmt.Errorf("TCP check failed: %w", err)
	}
	defer conn.Close()
	
	return nil
}

// checkGRPC performs a gRPC health check
func (c *Checker) checkGRPC(check *Check) error {
	// Placeholder for gRPC health check
	// In production, implement gRPC health check protocol
	c.log.Debug("gRPC health check not yet implemented")
	return nil
}

// generateCheckID generates a unique check ID
func generateCheckID() string {
	return fmt.Sprintf("check-%d", time.Now().UnixNano())
}

// SetCallback sets a callback function for a check
func (c *Checker) SetCallback(checkID string, callback func(status CheckStatus)) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	check, exists := c.checks[checkID]
	if !exists {
		return fmt.Errorf("check not found: %s", checkID)
	}
	
	check.callback = callback
	return nil
}
