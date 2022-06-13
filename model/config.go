package model

type Config struct {
	DB_HOST string `mapstructure:"DB_HOST"`
	DB_PORT string `mapstructure:"DB_PORT"`
	DB_NAME string `mapstructure:"DB_NAME"`
	DB_USER string `mapstructure:"DB_USER"`
	DB_PASS string `mapstructure:"DB_PASS"`

	RPC string `mapstructure:"RPC"`

	SCHEMA_DIR      string `mapstructure:"SCHEMA_DIR"`
	UPLOAD_CONTRACT string `mapstructure:"UPLOAD_CONTRACT"`

	SERVER string `mapstructure:"SERVER"`

	REDIS_HOST    string `mapstructure:"REDIS_HOST"`
	REDIS_PORT    string `mapstructure:"REDIS_PORT"`
	REDIS_CHANNEL string `mapstructure:"REDIS_CHANNEL"`
}
