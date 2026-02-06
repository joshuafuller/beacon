// Package main implements the mDNS bridge logic.
package main

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BridgeConfig defines bridge configuration.
type BridgeConfig struct {
	Interfaces      []string `yaml:"interfaces"`
	AllowedServices []string `yaml:"allowed_services"`
	ExcludeSubnets  []string `yaml:"exclude_subnets"`
}

// Bridge implements multi-interface mDNS forwarding.
type Bridge struct {
	config *BridgeConfig
}

// NewBridge creates a new bridge instance.
func NewBridge(config *BridgeConfig) (*Bridge, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if len(config.Interfaces) < 2 {
		return nil, fmt.Errorf("bridge requires at least 2 interfaces")
	}

	return &Bridge{
		config: config,
	}, nil
}

// Start begins bridging operations.
//
// This is a simplified educational implementation. A production bridge would:
// 1. Create per-interface querier/responder instances
// 2. Listen for mDNS queries on all interfaces
// 3. Filter queries based on AllowedServices
// 4. Forward matching queries to other interfaces
// 5. Rewrite IP addresses in responses (RFC 6762 §15)
// 6. Implement subnet exclusion logic
func (b *Bridge) Start(ctx context.Context) error {
	// Educational note: In production, this would:
	// - Set up per-interface listeners
	// - Start goroutines for each interface
	// - Implement query forwarding logic
	// - Handle context cancellation
	
	fmt.Println("NOTE: This is an educational example.")
	fmt.Println("Production bridging requires platform-specific interface binding.")
	fmt.Println("See F-10 Network Interface Management spec for implementation details.\n")

	return nil
}

// Stop halts bridge operations.
func (b *Bridge) Stop() error {
	// Educational note: In production, this would:
	// - Stop all interface listeners
	// - Close querier/responder instances
	// - Wait for goroutines to finish

	return nil
}

// LoadConfig loads bridge configuration from YAML file.
func LoadConfig(path string) (*BridgeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config BridgeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &config, nil
}

// isServiceAllowed checks if a service type is in the allowlist.
func (b *Bridge) isServiceAllowed(serviceType string) bool {
	for _, allowed := range b.config.AllowedServices {
		if allowed == serviceType {
			return true
		}
	}
	return false
}
