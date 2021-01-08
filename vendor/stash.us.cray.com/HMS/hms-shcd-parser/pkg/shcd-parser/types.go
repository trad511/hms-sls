package shcd_parser

type HMNRow struct {
	Source              string
	SourceRack          string
	SourceLocation      string
	SourceSubLocation   string `json:",omitempty"`
	SourceParent        string `json:",omitempty"`
	DestinationRack     string `json:",omitempty"`
	DestinationLocation string `json:",omitempty"`
	DestinationPort     string `json:",omitempty"`
}
