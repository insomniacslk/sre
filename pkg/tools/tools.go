package tools

var AllTools = map[string]Tool{
	// TODO this needs to be populated from the config file
}

type Tool struct {
	// Name returns the name of the tool
	Name string
	// URL returns the URL of the tool
	URL string
}
