package vpn

import (
	"errors"
	"fmt"

	"github.com/insomniacslk/sre/pkg/config"

	"github.com/mitchellh/go-homedir"
)

func NewVpn(cfg *config.VpnConfig) (*Vpn, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints specified")
	}
	if cfg.DefaultEndpoint == "" {
		return nil, fmt.Errorf("no default endpoint specified")
	}
	executable := "openconnect"
	if cfg.Executable != "" {
		executable = cfg.Executable
	}
	endpoints := make([]Endpoint, 0, len(cfg.Endpoints))
	for _, item := range cfg.Endpoints {
		endpoints = append(endpoints, Endpoint{Name: item["name"], Host: item["host"]})
	}
	pidFile := "~/.openconnect.pid"
	if cfg.PidFile != "" {
		pidFile = cfg.PidFile
	}
	pidFile, err := homedir.Expand(pidFile)
	if err != nil {
		return nil, fmt.Errorf("failed to expand pid_file path: %w", err)
	}
	return &Vpn{
		Executable:      executable,
		Endpoints:       endpoints,
		DefaultEndpoint: cfg.DefaultEndpoint,
		PidFile:         pidFile,
	}, nil
}

type Vpn struct {
	Executable      string
	Endpoints       []Endpoint
	DefaultEndpoint string
	PidFile         string
}

type Endpoint struct {
	Name string
	Host string
}

func (v *Vpn) Connect() error {
	/*
		similar to the following in bash:

			pid_file=~/.openconnect.pid
			eval $(openconnect --external-browser "<path to browser>" --authenticate your.vpn.example.org/path --useragent="AnyConnect <version>")
			[ -n "$COOKIE" ] && sudo openconnect \
			    $CONNECT_URL \
			    -b \
			    --servercert $FINGERPRINT \
			    --resolve $RESOLVE \
			    -C "$COOKIE" \
			    --useragent="AnyConnect whatever" \
			    --pid-file="${pid_file}"
	*/
	return errors.New("Vpn.Connect not implemented yet")
}

func (v *Vpn) Disconnect() error {
	return errors.New("Vpn.Disconnect not implemented yet")
}
