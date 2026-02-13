package cli

import (
	"log"

	"github.com/insomniacslk/sre/pkg/config"
	"github.com/insomniacslk/sre/pkg/vpn"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewVpnCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "vpn",
		Short: "Connect and disconnect from VPN using OpenConnect",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Debugf("Running vpn command")
			v, err := vpn.NewVpn(&cfg.Vpn)
			if err != nil {
				logrus.Fatalf("Failed to get VPN: %v", err)
			}
			subcmd := args[0]
			switch subcmd {
			case "connect":
				err = v.Connect()
			case "disconnect":
				err = v.Disconnect()
			default:
				logrus.Fatalf("Unknown subcommand %q", subcmd)
			}
			if err != nil {
				log.Fatalf("Subcommand %q failed: %v", subcmd, err)
			}
		},
	}
}
