package conf

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Turu struct {
	Config *Config `mapstructure:"config"`
}

type Config struct {
	ApisixYaml *ApisixYaml `mapstructure:"apisix-yaml"`
	ApisixEtcd *ApisixEtcd `mapstructure:"apisix-etcd"`
}

type MTLS struct {
	CA   string `mapstructure:"ca"`
	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
}

type ApisixYaml struct {
	Path string `mapstructure:"path"`
}

type ApisixEtcd struct {
	Endpoint []string      `mapstructure:"endpoint"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Username *string       `mapstructure:"username"`
	Password *string       `mapstructure:"password"`
	MTLS     *MTLS         `mapstructure:"mtls"`
}

var TuruConfig *Turu

func Init() {
	viper.AutomaticEnv()
	viper.SetConfigName("turu")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(viper.GetString("HOME"))
	viper.AddConfigPath("/etc/turu")
	viper.SetEnvPrefix("TURU")

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Msg(fmt.Sprint("Using config file:", viper.ConfigFileUsed()))
	}

	if err := viper.Unmarshal(&TuruConfig); err != nil {
		log.Fatal().Err(err).Stack().Msg("Failed to marshal config")
	}
}
