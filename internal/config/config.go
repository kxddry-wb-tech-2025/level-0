package config

type Config struct {
	Env     string  `yaml:"env" env-default:"dev"` // local, dev, prod
	Storage Storage `yaml:"storage"`
}
type Storage struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"require"`
}
