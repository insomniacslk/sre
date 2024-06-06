package tools

import (
	"fmt"

	"github.com/insomniacslk/sre/pkg/ansi"
	"github.com/insomniacslk/sre/pkg/config"
)

func List(cfg *config.ToolsConfig) error {
	if len(*cfg) == 0 {
		fmt.Printf("No tools specified in config\n")
		return nil
	}
	for _, tool := range *cfg {
		fmt.Printf("%s: %s\n", ansi.ToURL(tool.Name, tool.URL), tool.Description)
	}
	return nil
}
