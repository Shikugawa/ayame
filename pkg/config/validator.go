package config

import "fmt"

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
	for _, cfg := range configs {
		for _, device := range cfg.Devices {
			var includes bool

			for _, link := range linkConfigs {
				if link.Name == device.Name {
					includes = true
					break
				}
			}

			if !includes {
				return fmt.Errorf("unconfigured device found")
			}
		}
	}

	return nil
}
