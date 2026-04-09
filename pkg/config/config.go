// place for API server to load config, for now its the store's address

package config

type APIConfig struct {
	Address string
}

func DefaultAPIConfig() APIConfig {
	return APIConfig {
		Address: ":8089",
	}
}
