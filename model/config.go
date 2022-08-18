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
	WORKSPACE_DIR   string `mapstructure:"WORKSPACE_DIR"`
	ARTIFACTS       string `mapstructure:"ARTIFACTS"`

	SERVER string `mapstructure:"SERVER"`

	REDIS_HOST    string `mapstructure:"REDIS_HOST"`
	REDIS_PORT    string `mapstructure:"REDIS_PORT"`
	REDIS_CHANNEL string `mapstructure:"REDIS_CHANNEL"`

	AWS_REGION            string `mapstructure:"AWS_REGION"`
	AWS_ACCESS_KEY_ID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWS_SECRET_ACCESS_KEY string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	BUCKET_NAME           string `mapstructure:"BUCKET_NAME"`
	AWS_FOLDER            string `mapstructure:"AWS_FOLDER"`

	WORKSPACE_REGEX string `mapstructure:"WORKSPACE_REGEX"`

	ZIP_PREFIX string `mapstructure:"ZIP_PREFIX"`

	RUST_OPTIMIZER      string `mapstructure:"RUST_OPTIMIZER"`
	WORKSPACE_OPTIMIZER string `mapstructure:"WORKSPACE_OPTIMIZER"`

	DOCKER_USERNAME string `mapstructure:"DOCKER_USERNAME"`
	DOCKER_PASSWORD string `mapstructure:"DOCKER_PASSWORD"`
}
