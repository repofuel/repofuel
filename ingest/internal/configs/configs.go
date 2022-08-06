// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package configs

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/repofuel/repofuel/accounts/pkg/keys"
	"github.com/repofuel/repofuel/pkg/mongocon"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"github.com/repofuel/repofuel/pkg/utilconfig"
	"github.com/rs/zerolog/log"
)

type Configs struct {
	Keys     keys.ServiceKeys
	Repofuel repofuel.Options
	DB       mongocon.DatabaseOptions
	//deprecated
	Providers struct {
		//deprecated
		Github struct {
			AppID         int64              `yaml:"app_id"`
			PrivateKey    keys.RSAPrivateKey `yaml:"private_key"`
			Server        string             `yaml:"server"`
			WebhookSecret string             `yaml:"webhook_secret,omitempty"`
			AppName       string             `yaml:"app_name"`
		}
		//deprecated
		Jira struct {
			BaseURL  string `yaml:"base_url"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		}
	}
}

func Parse(ctx context.Context) (*Configs, error) {
	if err := godotenv.Load(".env", "ingest/.env"); err != nil {
		log.Ctx(ctx).Debug().Msg("missing .env file, will continue without it")
	}

	var cfg Configs
	if err := utilconfig.LoadYAMLFromEnvPath(&cfg, "COMMON_SECRETS"); err != nil {
		return nil, err
	}
	if err := utilconfig.LoadYAMLFromEnvPath(&cfg, "SERVICE_SECRETS"); err != nil {
		return nil, err
	}

	if key, ok := os.LookupEnv("PRIVATE_KEY_GITHUB"); ok {
		err := cfg.Providers.Github.PrivateKey.Load(key)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
