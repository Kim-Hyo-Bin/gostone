package password

// PasswordAuthRequest mirrors Identity API v3 auth JSON (password, token, optional scope).
// Keystone may send multiple methods; gostone currently supports a single method per request.
type PasswordAuthRequest struct {
	Auth struct {
		Identity struct {
			Methods  []string `json:"methods"`
			Password struct {
				User struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Password string `json:"password"`
					Domain   struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"domain"`
				} `json:"user"`
			} `json:"password"`
			Token *struct {
				ID string `json:"id"`
			} `json:"token"`
		} `json:"identity"`
		Scope *AuthScope `json:"scope"`
	} `json:"auth"`
}

// AuthScope is a subset of Keystone scope (project or domain).
type AuthScope struct {
	Project *struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Domain *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"domain"`
	} `json:"project"`
	Domain *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"domain"`
}
