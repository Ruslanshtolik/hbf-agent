package firewall

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/hbf-agent/internal/config"
)

// Manager manages firewall rules
type Manager struct {
	config    config.FirewallConfig
	log       *logrus.Logger
	backend   Backend
	rules     map[string]*Rule
	mu        sync.RWMutex
	stopChan  chan struct{}
	running   bool
}

// Backend represents a firewall backend (iptables or nftables)
type Backend interface {
	AddRule(rule *Rule) error
	DeleteRule(rule *Rule) error
	ListRules() ([]*Rule, error)
	Flush() error
	SetDefaultPolicy(chain, policy string) error
}

// Rule represents a firewall rule
type Rule struct {
	ID        string
	Chain     string
	Protocol  string
	Source    string
	Dest      string
	SPort     string
	DPort     string
	Action    string
	Comment   string
	CreatedAt time.Time
}

// NewManager creates a new firewall manager
func NewManager(cfg config.FirewallConfig, log *logrus.Logger) (*Manager, error) {
	var backend Backend
	var err error
	
	switch cfg.Backend {
	case "iptables":
		backend, err = NewIPTablesBackend(log)
	case "nftables":
		backend, err = NewNFTablesBackend(log)
	default:
		return nil, fmt.Errorf("unsupported firewall backend: %s", cfg.Backend)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall backend: %w", err)
	}
	
	return &Manager{
		config:   cfg,
		log:      log,
		backend:  backend,
		rules:    make(map[string]*Rule),
		stopChan: make(chan struct{}),
	}, nil
}

// Start starts the firewall manager
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("firewall manager is already running")
	}
	m.running = true
	m.mu.Unlock()
	
	m.log.Info("Starting firewall manager...")
	
	// Set default policies
	if err := m.setDefaultPolicies(); err != nil {
		return fmt.Errorf("failed to set default policies: %w", err)
	}
	
	// Load initial rules from config
	if err := m.loadConfigRules(); err != nil {
		return fmt.Errorf("failed to load config rules: %w", err)
	}
	
	// Start sync loop
	go m.syncLoop(ctx)
	
	return nil
}

// Stop stops the firewall manager
func (m *Manager) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return fmt.Errorf("firewall manager is not running")
	}
	m.running = false
	m.mu.Unlock()
	
	close(m.stopChan)
	m.log.Info("Firewall manager stopped")
	
	return nil
}

// AddRule adds a new firewall rule
func (m *Manager) AddRule(rule *Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if rule.ID == "" {
		rule.ID = generateRuleID()
	}
	rule.CreatedAt = time.Now()
	
	if err := m.backend.AddRule(rule); err != nil {
		return fmt.Errorf("failed to add rule: %w", err)
	}
	
	m.rules[rule.ID] = rule
	m.log.Infof("Added firewall rule: %s", rule.ID)
	
	return nil
}

// DeleteRule deletes a firewall rule
func (m *Manager) DeleteRule(ruleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	rule, exists := m.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}
	
	if err := m.backend.DeleteRule(rule); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	
	delete(m.rules, ruleID)
	m.log.Infof("Deleted firewall rule: %s", ruleID)
	
	return nil
}

// ListRules returns all firewall rules
func (m *Manager) ListRules() []*Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	rules := make([]*Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		rules = append(rules, rule)
	}
	
	return rules
}

// GetRule returns a specific rule by ID
func (m *Manager) GetRule(ruleID string) (*Rule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	rule, exists := m.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}
	
	return rule, nil
}

// Flush removes all firewall rules
func (m *Manager) Flush() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if err := m.backend.Flush(); err != nil {
		return fmt.Errorf("failed to flush rules: %w", err)
	}
	
	m.rules = make(map[string]*Rule)
	m.log.Info("Flushed all firewall rules")
	
	return nil
}

// setDefaultPolicies sets the default firewall policies
func (m *Manager) setDefaultPolicies() error {
	chains := []string{"INPUT", "FORWARD", "OUTPUT"}
	policy := "ACCEPT"
	
	if m.config.DefaultPolicy == "deny" {
		policy = "DROP"
	}
	
	for _, chain := range chains {
		if err := m.backend.SetDefaultPolicy(chain, policy); err != nil {
			return fmt.Errorf("failed to set default policy for %s: %w", chain, err)
		}
	}
	
	m.log.Infof("Set default policy to %s", policy)
	return nil
}

// loadConfigRules loads rules from configuration
func (m *Manager) loadConfigRules() error {
	for _, cfgRule := range m.config.Rules {
		rule := &Rule{
			Chain:    cfgRule.Chain,
			Protocol: cfgRule.Protocol,
			Source:   cfgRule.Source,
			Dest:     cfgRule.Dest,
			SPort:    cfgRule.SPort,
			DPort:    cfgRule.DPort,
			Action:   cfgRule.Action,
			Comment:  cfgRule.Comment,
		}
		
		if err := m.AddRule(rule); err != nil {
			m.log.Errorf("Failed to add config rule: %v", err)
		}
	}
	
	return nil
}

// syncLoop periodically syncs firewall rules
func (m *Manager) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			if err := m.sync(); err != nil {
				m.log.Errorf("Failed to sync firewall rules: %v", err)
			}
		}
	}
}

// sync synchronizes firewall rules with the backend
func (m *Manager) sync() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	backendRules, err := m.backend.ListRules()
	if err != nil {
		return fmt.Errorf("failed to list backend rules: %w", err)
	}
	
	// Check for missing rules and add them
	for _, rule := range m.rules {
		found := false
		for _, backendRule := range backendRules {
			if rulesEqual(rule, backendRule) {
				found = true
				break
			}
		}
		
		if !found {
			m.log.Warnf("Rule %s missing from backend, re-adding", rule.ID)
			if err := m.backend.AddRule(rule); err != nil {
				m.log.Errorf("Failed to re-add rule %s: %v", rule.ID, err)
			}
		}
	}
	
	return nil
}

// rulesEqual checks if two rules are equal
func rulesEqual(r1, r2 *Rule) bool {
	return r1.Chain == r2.Chain &&
		r1.Protocol == r2.Protocol &&
		r1.Source == r2.Source &&
		r1.Dest == r2.Dest &&
		r1.SPort == r2.SPort &&
		r1.DPort == r2.DPort &&
		r1.Action == r2.Action
}

// generateRuleID generates a unique rule ID
func generateRuleID() string {
	return fmt.Sprintf("rule-%d", time.Now().UnixNano())
}

// IPTablesBackend implements the Backend interface using iptables
type IPTablesBackend struct {
	ipt *iptables.IPTables
	log *logrus.Logger
}

// NewIPTablesBackend creates a new iptables backend
func NewIPTablesBackend(log *logrus.Logger) (*IPTablesBackend, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize iptables: %w", err)
	}
	
	return &IPTablesBackend{
		ipt: ipt,
		log: log,
	}, nil
}

// AddRule adds a rule using iptables
func (b *IPTablesBackend) AddRule(rule *Rule) error {
	ruleSpec := b.buildRuleSpec(rule)
	
	if err := b.ipt.AppendUnique("filter", rule.Chain, ruleSpec...); err != nil {
		return fmt.Errorf("failed to add iptables rule: %w", err)
	}
	
	return nil
}

// DeleteRule deletes a rule using iptables
func (b *IPTablesBackend) DeleteRule(rule *Rule) error {
	ruleSpec := b.buildRuleSpec(rule)
	
	if err := b.ipt.Delete("filter", rule.Chain, ruleSpec...); err != nil {
		return fmt.Errorf("failed to delete iptables rule: %w", err)
	}
	
	return nil
}

// ListRules lists all rules using iptables
func (b *IPTablesBackend) ListRules() ([]*Rule, error) {
	// This is a simplified implementation
	// In production, you'd parse iptables-save output
	return []*Rule{}, nil
}

// Flush flushes all rules using iptables
func (b *IPTablesBackend) Flush() error {
	chains := []string{"INPUT", "FORWARD", "OUTPUT"}
	
	for _, chain := range chains {
		if err := b.ipt.ClearChain("filter", chain); err != nil {
			return fmt.Errorf("failed to clear chain %s: %w", chain, err)
		}
	}
	
	return nil
}

// SetDefaultPolicy sets the default policy for a chain
func (b *IPTablesBackend) SetDefaultPolicy(chain, policy string) error {
	if err := b.ipt.ChangePolicy("filter", chain, policy); err != nil {
		return fmt.Errorf("failed to set policy: %w", err)
	}
	
	return nil
}

// buildRuleSpec builds an iptables rule specification
func (b *IPTablesBackend) buildRuleSpec(rule *Rule) []string {
	spec := []string{}
	
	if rule.Protocol != "" {
		spec = append(spec, "-p", rule.Protocol)
	}
	
	if rule.Source != "" {
		spec = append(spec, "-s", rule.Source)
	}
	
	if rule.Dest != "" {
		spec = append(spec, "-d", rule.Dest)
	}
	
	if rule.SPort != "" {
		spec = append(spec, "--sport", rule.SPort)
	}
	
	if rule.DPort != "" {
		spec = append(spec, "--dport", rule.DPort)
	}
	
	if rule.Comment != "" {
		spec = append(spec, "-m", "comment", "--comment", rule.Comment)
	}
	
	spec = append(spec, "-j", rule.Action)
	
	return spec
}

// NFTablesBackend implements the Backend interface using nftables
type NFTablesBackend struct {
	log *logrus.Logger
}

// NewNFTablesBackend creates a new nftables backend
func NewNFTablesBackend(log *logrus.Logger) (*NFTablesBackend, error) {
	// This is a placeholder implementation
	// In production, you'd use github.com/google/nftables
	return &NFTablesBackend{
		log: log,
	}, nil
}

// AddRule adds a rule using nftables
func (b *NFTablesBackend) AddRule(rule *Rule) error {
	// Placeholder implementation
	b.log.Infof("Adding nftables rule: %+v", rule)
	return nil
}

// DeleteRule deletes a rule using nftables
func (b *NFTablesBackend) DeleteRule(rule *Rule) error {
	// Placeholder implementation
	b.log.Infof("Deleting nftables rule: %+v", rule)
	return nil
}

// ListRules lists all rules using nftables
func (b *NFTablesBackend) ListRules() ([]*Rule, error) {
	// Placeholder implementation
	return []*Rule{}, nil
}

// Flush flushes all rules using nftables
func (b *NFTablesBackend) Flush() error {
	// Placeholder implementation
	b.log.Info("Flushing nftables rules")
	return nil
}

// SetDefaultPolicy sets the default policy for a chain
func (b *NFTablesBackend) SetDefaultPolicy(chain, policy string) error {
	// Placeholder implementation
	b.log.Infof("Setting nftables default policy: %s -> %s", chain, policy)
	return nil
}
