package main

import (
	_ "embed"

	"github.com/sirupsen/logrus"

	"github.com/insomniacslk/sre/pkg/cli"
)

const progname = "sre"

func main() {
	cfg, err := cli.InitConfig(progname)
	if err != nil {
		logrus.Fatalf("Failed to get configuration: %v", err)
	}
	rootCmd, err := cli.WithSubcommands(
		progname,
		cli.NewConfigExampleCmd(cfg),
		cli.NewShowConfigCmd(cfg),
		cli.NewToolsCmd(cfg),
		cli.NewOmgCmd(cfg),
		cli.NewOmgTemplateCmd(cfg),
		cli.NewVpnCmd(cfg),
		cli.OncallCmd,
		cli.NewNotificationsCmd(cfg),
		cli.NewIncidentsCmd(cfg),
	)
	if err != nil {
		logrus.Fatalf("Failed to initialize root command: %v", err)
	}
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("Failed to execute command: %v", err)
	}
}
