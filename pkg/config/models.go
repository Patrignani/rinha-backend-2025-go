package config

import "time"

type Environment struct {
	Postgres       Postgres
	StartPort      string `env:"START_PORT,default=8080"`
	Queue          Queue
	DefaultUrl     string        `env:"DEFAULT_URL"`
	FallbackUrl    string        `env:"FALLBACK_URL"`
	AttempsRetry   int           `env:"ATTEMPS_RETRY"`
	TimeAttemps    time.Duration `env:"TIME_ATTEMPS"`
	UseQueueInPost bool          `env:"USE_QUEUE_IN_POST,default=false"`
}

type Queue struct {
	Buffer  int `env:"QUEUE_BUFFER"`
	Workers int `env:"QUEUE_WORKERS"`
}

type Postgres struct {
	Host string `env:"DB_HOST"`
	User string `env:"DB_USER"`
	Pass string `env:"DB_PASSWORD"`
	Name string `env:"DB_NAME"`
	PORT string `env:"DB_PORT,default=5432"`
}
