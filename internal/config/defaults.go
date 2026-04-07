package config

func defaultConfig() *Config {
	c := &Config{}
	c.Database.Connection = "file::memory:?cache=shared"
	c.Service.Listen = ":8080"
	c.Token.ExpirationHours = 24
	return c
}
