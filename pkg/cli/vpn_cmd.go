package cli

import (
	"insomniac/sre/pkg/config"
	"insomniac/sre/pkg/vpn"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewVpnCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "vpn",
		Short: "Connect and disconnect from VPN using OpenConnect",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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
