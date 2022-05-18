package configs

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/repofuel/repofuel/accounts/pkg/keys"
	"github.com/repofuel/repofuel/pkg/mongocon"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/repofuel/repofuel/pkg/utilconfig"
)

type Configs struct {
	Keys     keys.ServiceKeys
	Repofuel repofuel.Options
	DB       mongocon.DatabaseOptions
}

func Parse() (*Configs, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("missing .env file, will continue without it")
	}

	var cfg Configs
	if err := utilconfig.LoadYAMLFromEnvPath(&cfg, "COMMON_SECRETS"); err != nil {
		return nil, err
	}
	if err := utilconfig.LoadYAMLFromEnvPath(&cfg, "SERVICE_SECRETS"); err != nil {
		return nil, err
	}

	return &cfg, nil
}
