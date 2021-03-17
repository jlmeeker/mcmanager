package releases

// FLAVORS is a list of our supported server flavors
var FLAVORS = []string{"vanilla", "spigot", "paper"}

// Version is the (important) fields for each release
type Version struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	ReleaseTime string `json:"releaseTime"`
	Build       int    `json:"build"`
}

// VersionFile is the structure of the version_manifest.json file
type VersionFile struct {
	Latest struct {
		Release  string
		Snapshot string
	}
	Versions []Version
}

// FlavorIsValid returns if a given flavor is in the list of FLAVORS
func FlavorIsValid(flavor string) bool {
	for _, f := range FLAVORS {
		if f == flavor {
			return true
		}
	}
	return false
}
