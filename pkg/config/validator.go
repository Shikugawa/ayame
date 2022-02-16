package config

import "fmt"

func ValidateLinkConfigs(linkConfigs []LinkConfig) error {
	// Check required fields
	for _, cfg := range linkConfigs {
		if cfg.LinkMode == "" {
			return fmt.Errorf("LinkMode must not be empty")
		}
		if cfg.Name == "" {
			return fmt.Errorf("Name must not be empty")
		}
	}

	// Check duplicate of names
	tmp := make(map[string]bool)
	for _, cfg := range linkConfigs {
		if _, ok := tmp[cfg.Name]; ok {
			return fmt.Errorf("Name must not be duplicated")
		}

		tmp[cfg.Name] = false
	}

	return nil
}

func ValidateNamespace(configs []NamespaceConfig, linkConfigs []LinkConfig) error {
	// Unique Name
	tmp := make(map[string]bool)
	for _, cfg := range configs {
		if _, ok := tmp[cfg.Name]; ok {
			return fmt.Errorf("namespace name must be unique")
		}
		tmp[cfg.Name] = false
	}

	// Device Exists
	deviceNameContainsInLink := func(name string) bool {
		for _, link := range linkConfigs {
			if name == link.Name {
				return true
			}
		}
		return false
	}

	for _, cfg := range configs {
		for _, device := range cfg.Devices {
			if !deviceNameContainsInLink(device.Name) {
				return fmt.Errorf("unconfigured device found")
			}
		}
	}

	return nil
}
