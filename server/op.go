package server

// Op is the structure of an op within the ops.json file
type Op struct {
	UUID              string `json:"uuid"`
	Name              string `json:"name"`
	Level             int8   `json:"level"`
	BypassPlayerLimit bool   `json:"bypassesPlayerLimit"`
}

// WLPlayer is the structure of a payer within the whitelist.json file
type WLPlayer struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
