package types

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	// OutputFormatPretty is the default human-readable output format
	OutputFormatPretty OutputFormat = "pretty"
	// OutputFormatJSON is the JSON output format
	OutputFormatJSON OutputFormat = "json"
)

// String returns the string representation of the output format
func (f OutputFormat) String() string {
	return string(f)
}

// IsValid checks if the output format is valid
func (f OutputFormat) IsValid() bool {
	switch f {
	case OutputFormatPretty, OutputFormatJSON:
		return true
	default:
		return false
	}
}

// SupportedOutputFormats returns a list of all supported output formats
func SupportedOutputFormats() []OutputFormat {
	return []OutputFormat{
		OutputFormatPretty,
		OutputFormatJSON,
	}
}

// OutputMetadata contains metadata about the check execution
type OutputMetadata struct {
	DateTime string `json:"datetime"`
	Version  string `json:"version"`
	OS       string `json:"os"`
}

// JSONOutput represents the full JSON output format including results and metadata
type JSONOutput struct {
	Results  []CheckResult  `json:"results"`
	Metadata OutputMetadata `json:"metadata"`
}
