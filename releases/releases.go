package releases

// Version is the (important) fields for each release in the version_manifest.json file
type Version struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	ReleaseTime string `json:"releaseTime"`
}

// VersionFile is the structure of the version_manifest.json file
type VersionFile struct {
	Latest struct {
		Release  string
		Snapshot string
	}
	Versions []Version
}
