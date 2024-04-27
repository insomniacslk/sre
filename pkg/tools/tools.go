package tools

var AllTools = map[string]Tool{
	// TODO this needs to be populated from the config file
}

type Tool interface {
	// Name returns the name of the tool
	Name() string
	// IsInstalled returns whether the tool is installed in the provided
	// installDir
	IsInstalled(installDir string) (bool, error)
	// Install installs the tool in installDir, fetching it into srcDir
	Install(installDir, srcDir string) (string, error)
	// Update updates the tool from srcDir
	Update(installDir, srcDir string) (string, error)
}
