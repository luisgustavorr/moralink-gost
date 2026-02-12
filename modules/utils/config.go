package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func ConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(base, "moralingost")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", err
	}

	return path, nil
}

func LoadConfig() {
	cfgDir, err := ConfigPath()
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	fmt.Println("✅ 🔧 Creating config file in : ", cfgDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// First run: create default config
			viper.Set("server.grpc_port", 50051)
			viper.Set("api.user", "teste_disparo_shark")

			configFile := filepath.Join(cfgDir, "config.yaml")
			if err := viper.WriteConfigAs(configFile); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}
