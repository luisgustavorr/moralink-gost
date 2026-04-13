package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

func ConfigPath() string {
	// Priority 1: explicit env override (great for systemd with Environment= directive)
	if override := os.Getenv("MORALINK_CONFIG_DIR"); override != "" {
		_ = os.MkdirAll(override, 0o755)
		return override
	}

	// Priority 2: try the standard user config dir
	if base, err := os.UserConfigDir(); err == nil && base != "" {
		path := filepath.Join(base, "moralink-gost")
		if err := os.MkdirAll(path, 0o755); err == nil {
			return path
		}
	}

	// Priority 3: OS-specific system-level fallback (for services without $HOME)
	var systemPath string
	if runtime.GOOS == "windows" {
		// PROGRAMDATA is always set on Windows, even for SYSTEM account
		programData := os.Getenv("PROGRAMDATA")
		if programData == "" {
			programData = `C:\ProgramData`
		}
		systemPath = filepath.Join(programData, "moralink-gost")
	} else {
		systemPath = "/etc/moralink-gost"
	}

	if err := os.MkdirAll(systemPath, 0o755); err != nil {
		// Priority 4: last resort — next to the executable
		exe, err2 := os.Executable()
		if err2 != nil {
			log.Fatalf("cannot determine config path: %v", err)
		}
		systemPath = filepath.Dir(exe)
		log.Printf("⚠️  Warning: using executable directory as config path: %s", systemPath)
	}

	return systemPath
}

func LoadConfig() {
	cfgDir := ConfigPath()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	log.Println("✅ 🔧 Creating config file in nv : ", cfgDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// First run: create default config
			viper.Set("server.grpc_port", 50051)
			viper.Set("api.user", "default")
			viper.Set("api.mode", "prod")
			viper.Set("api.token", "1234567")
			configFile := filepath.Join(cfgDir, "config.yaml")
			if err := viper.WriteConfigAs(configFile); err != nil {
				log.Fatalf("failed to write default config to %s: %v", configFile, err)
			}
			log.Println("✅ 🔧 Default config written to:", configFile)
		} else {
			log.Fatalf("failed to read config: %v", err)
		}
	}
}
