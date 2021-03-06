package forms

// NewServer is the expected of the data expected from the new server web form
type NewServer struct {
	Release   string `form:"release"`
	AutoStart bool   `form:"autostart"`
	Flavor    string `form:"flavor"`
	MOTD      string `form:"motd"`
	Name      string `form:"name"`
	Page      string `form:"page"`
	StartNow  bool   `form:"startnow"`
	Whitelist bool   `form:"whitelist"`
}

// Login is the structure of the data expected from the login web form
type Login struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Page     string `form:"page"`
}

// AddOp is the structure of the data extected from the addop web form
type AddOp struct {
	OpName string `form:"opname"`
}
