package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"os"
)

type ServiceConfig struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"prettyjson"`
	Port string `env:"PORT" envDefault:"8080"`

	DatabaseHost string `env:"DB_HOST"`
	DatabaseName string `env:"DB_NAME" envDefault:"af_names"`
	DatabasePort int `env:"DB_PORT" envDefault:"3306"`
	DatabaseUser string `env:"DB_USER" envDefault:"af_namer"`
	DatabasePass string `env:"DB_PASS"`
	DatabaseDriver string `env:"DB_DRIVER" envDefault:"mysql"`
	db *sqlx.DB
}
//namer$006

func (cfg *ServiceConfig) getDbDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.DatabaseUser,
		cfg.DatabasePass,
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName)
}
func (cfg *ServiceConfig) GetDbConn() (*sqlx.DB, error) {
	if cfg.db == nil {
		log.WithField("dsn", cfg.getDbDSN()).Debug("Connecting to db")
		db, err := sqlx.Connect(cfg.DatabaseDriver, cfg.getDbDSN())
		if err != nil {
			return nil, err
		}
		cfg.db = db
	}
	return cfg.db, nil
}

var singletonConfig *ServiceConfig
func LoadConfig() *ServiceConfig {
	if singletonConfig == nil {
		cfg := new(ServiceConfig)
		err := env.Parse(cfg)
		if err != nil {
			log.WithError(err).Fatalf("Error loading main env configs")
			os.Exit(1)
		}

		singletonConfig = cfg
	}
	return singletonConfig
}