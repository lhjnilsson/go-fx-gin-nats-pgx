package config

type ApplicationConfig struct {
	DatabaseURL     string
	NatsURL         string
	ServerAddress   string
	AlpacaAPIKey    string
	AlpacaAPISecret string
	AlpacaAPIURL    string
}

func New() *ApplicationConfig {
	return &ApplicationConfig{
		DatabaseURL:     "postgres://postgres:password@localhost:5432/postgres?sslmode=disable",
		NatsURL:         "nats://localhost:4222",
		ServerAddress:   ":8080",
		AlpacaAPIKey:    "",
		AlpacaAPISecret: "",
		AlpacaAPIURL:    "https://paper-api.alpaca.markets"}
}
