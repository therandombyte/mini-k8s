// place for API server to load config, for now its the store's address

package config

type APIConfig struct {
	Address string
}

func DefautAPIConfig() APIConfig {
	return APIConfig {
		Address: ":8080",
	}
}
