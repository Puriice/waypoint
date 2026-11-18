package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/puriice/waypoint/domain/unit"
)

type Config struct {
	JWT          jwtPolicy
	LoginPenalty LoginPenalty `json:"login_penalty"`
}

var ConfigPath = "./config.json"
var config *Config = &Config{}
var once sync.Once

func Default() Config {
	return Config{
		JWT: jwtPolicy{
			Issuer: "purinutt.com",
			TTL: ttl{
				AccessToken:  unit.Seconds(5 * time.Minute),
				RefreshToken: unit.Seconds(30 * 24 * time.Hour),
			},
		},
		LoginPenalty: LoginPenalty{
			ForgivenWindow: 3,
			Penalties: []unit.Seconds{
				unit.Seconds(time.Minute),
				unit.Seconds(time.Minute),
				unit.Seconds(5 * time.Minute),
				unit.Seconds(5 * time.Minute),
				unit.Seconds(5 * time.Minute),
				unit.Seconds(5 * time.Minute),
				unit.Seconds(5 * time.Minute),
				unit.Seconds(30 * time.Minute),
				unit.Seconds(30 * time.Minute),
				unit.Seconds(30 * time.Minute),
				unit.Seconds(30 * time.Minute),
				unit.Seconds(30 * time.Minute),
				unit.Seconds(time.Hour),
			},
			RequireOTP: true,
		},
	}
}

func load() {
	content, err := os.ReadFile(ConfigPath)

	if err == nil {
		if err := json.Unmarshal(content, config); err != nil {
			*config = Default()
		}
	} else {
		*config = Default()
	}
}

func Get() *Config {
	once.Do(load)

	return config
}
