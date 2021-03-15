package server

type Permission struct {
	Name           string `json:"name"`
	Allowed        bool   `json:"allowed"`
	RequireRunning bool   `json:"reqRunning"`
}

type Permissions map[string]Permission

func newPermissions() Permissions {
	var p = make(Permissions)
	p["ado"] = Permission{Name: "Add Op", RequireRunning: true}
	p["adw"] = Permission{Name: "Add Whitelist", RequireRunning: true}
	p["bkp"] = Permission{Name: "Backup"}
	p["day"] = Permission{Name: "Set Time Day", RequireRunning: true}
	p["sav"] = Permission{Name: "Save", RequireRunning: true}
	p["wea"] = Permission{Name: "Weather Clear", RequireRunning: true}
	p["del"] = Permission{Name: "Delete"}
	p["rgn"] = Permission{Name: "Regen World"}
	p["sta"] = Permission{Name: "Start"}
	p["sto"] = Permission{Name: "Stop", RequireRunning: true}
	return p
}

func PermissionsOp() Permissions {
	var allowed = []string{
		"ado",
		"adw",
		"bkp",
		"day",
		"sav",
		"wea",
	}

	return allowPerms(allowed, newPermissions())
}

func PermissionsOwner() Permissions {
	var allowed = []string{
		"del",
		"rgn",
		"sta",
		"sto",
	}

	return allowPerms(allowed, PermissionsOp())
}

func PermissionsPlayer() Permissions {
	return newPermissions()
}

func allowPerms(allowed []string, p Permissions) Permissions {
	for _, key := range allowed {
		perm := p[key]
		perm.Allowed = true
		p[key] = perm
	}
	return p
}
