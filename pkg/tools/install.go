package tools

import (
	"fmt"
	"insomniac/sre/pkg/config"

	"github.com/sirupsen/logrus"
)

func Install(cfg *config.ToolsConfig, names []string) error {
	toInstall := make(map[string]*Tool, 0)
	// remove duplicates and validate that the tools exist
	for _, name := range names {
		tool, found := AllTools[name]
		if !found {
			return fmt.Errorf("Unknown tool %q", name)
		}
		toInstall[name] = &tool
	}
	if len(toInstall) == 0 {
		return fmt.Errorf("No tool to install")
	}
	for name, tool := range toInstall {
		path, err := (*tool).Install(cfg.InstallDir, cfg.SrcDir)
		if err != nil {
			return fmt.Errorf("failed to install tool %q: %q", name, err)
		}
		logrus.Infof("Installed tool %q to %q", name, path)
	}
	return nil
}
