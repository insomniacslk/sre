package tools

import (
	"fmt"
	"insomniac/sre/pkg/config"

	"github.com/sirupsen/logrus"
)

func Update(cfg *config.ToolsConfig, names []string) error {
	toUpdate := make(map[string]*Tool, 0)
	// remove duplicates and validate that the tools exist
	for _, name := range names {
		tool, found := AllTools[name]
		if !found {
			return fmt.Errorf("Unknown tool %q", name)
		}
		toUpdate[name] = &tool
	}
	if len(toUpdate) == 0 {
		return fmt.Errorf("No tool to update")
	}
	for name, tool := range toUpdate {
		path, err := (*tool).Update(cfg.InstallDir, cfg.SrcDir)
		if err != nil {
			return fmt.Errorf("failed to update tool %q: %q", name, err)
		}
		logrus.Infof("Updated tool %q to %q", name, path)
	}
	return nil
}
