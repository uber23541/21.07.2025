package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port              string   `env:"PORT" env-default:"8080"`
	AllowedExtensions []string `env:"ALLOWED_EXTENSIONS" env-separator:"," env-default:".pdf,.jpeg"`
	MaxTasks          int      `env:"MAX_TASKS" env-default:"3"`
	MaxFilesPerTask   int      `env:"MAX_FILES_PER_TASK" env-default:"3"`
	StorageDir        string   `env:"STORAGE_DIR" env-default:"../storage"`
}

func LoadConfig() Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("config load: %v", err)
	}
	return cfg
}
