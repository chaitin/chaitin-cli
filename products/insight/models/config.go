package models

// Config holds the product-specific configuration.
type Config struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}
