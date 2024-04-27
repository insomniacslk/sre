package tools

import (
	"fmt"
	"insomniac/sre/pkg/config"
)

func List(cfg *config.ToolsConfig) error {
	for name, tool := range AllTools {
		isInstalled, err := tool.IsInstalled(cfg.InstallDir)
		if err != nil {
			return fmt.Errorf("failed to check if %q is installed: %w", name, err)
		}
		if isInstalled {
			fmt.Printf("✅ %s\n", name)
		} else {
			fmt.Printf("   %s\n", name)
		}
	}
	return nil
}
