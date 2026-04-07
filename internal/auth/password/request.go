package password

// PasswordAuthRequest mirrors the Identity API password auth JSON (subset).
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
		} `json:"identity"`
	} `json:"auth"`
}
