package conf

func defaultConfig() *Config {
	c := &Config{}
	c.Database.Connection = "file::memory:?cache=shared"
	c.Service.Listen = ":5000"
	c.Token.ExpirationHours = 24
	return c
}
