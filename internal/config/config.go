package config

import (
	"fmt"
	"github.com/caarlos0/env/v9"
	"sync"
	"time"
)

type Config struct {
	Production bool `env:"PRODUCTION"`

	LoginTTL    time.Duration `env:"LOGIN_TTL" envDefault:"720h"` // ~ 1 month
	AlertOnInit bool          `env:"ALERT_ON_INIT"`

	HttpPort  int    `env:"HTTP_PORT" envDefault:"8080"`
	BadgerURI string `env:"BADGER_URI" envDefault:"./badger"`

	IpApiEnable bool   `env:"IP_API_ENABLE" envDefault:"true"`
	IpApiKey    string `env:"IP_API_KEY"`

	AlertFromEmail     string `env:"ALERT_EMAIL_FROM"`
	AlertEmailTitle    string `env:"ALERT_EMAIL_TITLE" envDefault:"[{{.Service}}]: Your {{.Service}} account has been accessed from a new IP Address"`
	AlertEmailTemplate string `env:"ALERT_EMAIL_TEMPLATE,file"`

	MMailerURL string `env:"MMAILER_URL"`
	MMailerKey string `env:"MMAILER_KEY"`

	SMTPServer   string `env:"SMTP_SERVER"`
	SMTPPort     int    `env:"SMTP_PORT"`
	SMTPUser     string `env:"SMTP_USER"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
}

var cfg Config
var once sync.Once

func Get() Config {
	once.Do(func() {
		if err := env.Parse(&cfg); err != nil {
			panic(err)
		}
		if cfg.AlertEmailTemplate == "" {
			cfg.AlertEmailTemplate = DefaultEmailTemplate
		}

		fmt.Println(cfg.LoginTTL)

	})

	return cfg
}
