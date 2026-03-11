package app

import (
	"gopkg.in/ini.v1"
)

type MeasurementAPIInfo struct {
}

type SyncAPIInfo struct {
	Host string
}

type Env struct {
	Name                                 string
	AdminHost                            string
	MeasureServerHost                    string
	MeasureServerAPIKey                  string
	MeasureServerAmzCustomauthorizerName string
	SyncServerHost                       string
	SyncServerAPIKey                     string
}

const (
	Production = "Production"
	Staging    = "Staging"
)

var envs = map[string]Env{
	// 本番環境
	Production: {
		Name:                                 Production,
		AdminHost:                            "https://www.ecoramdar.jp",
		MeasureServerHost:                    "https://atjgvl5dieils-ats.iot.ap-northeast-1.amazonaws.com/topics/prod/frig",
		MeasureServerAPIKey:                  "TsZytYaZYnWSwK88vFKLiMnRbHyfj8wJuDQRietMY4VofKe9UyUZeYDZTimyUJQT",
		MeasureServerAmzCustomauthorizerName: "frig-staging-authorizer",
		SyncServerHost:                       "https://api.ecoramdar.jp",
		SyncServerAPIKey:                     "qcf9Q2zwLF7P8T35zWzL5Z9Fq5uKM8dehMtFmwHhPD4P5wu7ySZTg2McWmip7zsN",
	},

	// ステージング環境
	Staging: {
		Name:                                 Staging,
		AdminHost:                            "https://staging.www.ecoramdar.jp",
		MeasureServerHost:                    "https://atjgvl5dieils-ats.iot.ap-northeast-1.amazonaws.com/topics/staging/frig",
		MeasureServerAPIKey:                  "3vgapMkB8Q27nepB8Sj8ghFQ6bYamh7y5zbqRe9imfaibJQZiPKdFFD75i6WCpxA",
		MeasureServerAmzCustomauthorizerName: "frig-staging-authorizer",
		SyncServerHost:                       "https://staging.api.ecoramdar.jp",
		SyncServerAPIKey:                     "tCFXMdHETy5rSECniAhKm6NB3tkv9QZjugBLgGfghLfPZ4sZai4YoSTcMJkZepvP",
	},
}

const configPath_A9E = "/vol_data/config.ini"
const configPath_G3L = "/home/astina/config.ini"

func GetEnv(envType string) (Env, error) {
	env, ok := envs[envType]
	if !ok {
		panic("EnvType is unknown")
	}

	return env, nil
}

func ReadEnv() (Env, error) {
	var env Env

	configPath := ""
	if MODEL == "A9E" {
		configPath = configPath_A9E
	} else {
		configPath = configPath_G3L
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		Logger.Writef(LOG_LEVEL_ERR, "File Read Failed %s", err.Error())
		env, err = GetEnv(ENV)
		//return env, err
	} else {
		env = Env{
			Name:                                 cfg.Section("env").Key("name").String(),
			AdminHost:                            cfg.Section("env").Key("adminhost").String(),
			MeasureServerHost:                    cfg.Section("env").Key("measureserverhost").String(),
			MeasureServerAPIKey:                  cfg.Section("env").Key("measureserverapikey").String(),
			MeasureServerAmzCustomauthorizerName: cfg.Section("env").Key("measureserveramzcustomauthorizername").String(),
			SyncServerHost:                       cfg.Section("env").Key("syncserverhost").String(),
			SyncServerAPIKey:                     cfg.Section("env").Key("syncserverapikey").String(),
		}
	}
	Logger.Writef(LOG_LEVEL_INFO, "env = %s\n", env.Name)
	return env, err
}
