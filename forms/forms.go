package forms

// NewServer is the expected of the data expected from the new server web form
type NewServer struct {
	Name      string `form:"name"`
	MOTD      string `form:"motd"`
	Flavor    string `form:"flavor"`
	Release   string `form:"release"`
	AutoStart bool   `form:"autostart"`
	StartNow  bool   `form:"startnow"`
	Page      string `form:"page"`
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
