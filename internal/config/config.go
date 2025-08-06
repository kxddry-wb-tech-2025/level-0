package config

import "time"

type Config struct {
	Env     string  `yaml:"env" env-default:"dev"` // local, dev, prod
	Storage Storage `yaml:"storage"`
	Kafka   Kafka   `yaml:"kafka"`
}
type Storage struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"require"`
}

type Kafka struct {
	Brokers []string     `yaml:"brokers" env-required:"true"`
	Reader  ReaderConfig `yaml:"reader" env-required:"true"`
	Writer  WriterConfig `yaml:"writer" env-required:"true"`
}

type ReaderConfig struct {
	Topic          string        `yaml:"topic" env-required:"true"`
	GroupID        string        `yaml:"group_id" env-required:"true"`
	MinBytes       int           `yaml:"min_bytes" env-default:"1"`         // min fetch bytes
	MaxBytes       int           `yaml:"max_bytes" env-default:"1048576"`   // 1MB
	CommitInterval time.Duration `yaml:"commit_interval" env-default:"1s"`  // time.Duration, e.g. 1s
	StartOffset    string        `yaml:"start_offset" env-default:"latest"` // earliest | latest
}

type WriterConfig struct {
	Topic           string        `yaml:"topic" env-required:"true"`
	ClientID        string        `yaml:"client_id" env-required:"true"`
	Retries         int           `yaml:"retries" env-default:"5"`
	MaxMessageBytes int           `yaml:"max_message_bytes" env-default:"1048576"`
	Acks            string        `yaml:"acks" env-default:"all"`        // 0 | 1 | all
	Compression     string        `yaml:"compression" env-default:"lz4"` // lz4 | snappy | none | gzip | zstd
	Timeout         time.Duration `yaml:"timeout" env-default:"5s"`      // time.Duration
}
