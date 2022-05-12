package model

type Config struct {
	DB_HOST string `mapstructure:"DB_HOST"`
	DB_PORT string `mapstructure:"DB_PORT"`
	DB_NAME string `mapstructure:"DB_NAME"`
	DB_USER string `mapstructure:"DB_USER"`
	DB_PASS string `mapstructure:"DB_PASS"`

	RPC string `mapstructure:"RPC"`

	DIR             string `mapstructure:"DIR"`
	UPLOAD_CONTRACT string `mapstructure:"UPLOAD_CONTRACT"`
}
