package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func getAbsPath(path string) string {
	path = os.ExpandEnv(path)
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	return path
}

type (
	// Config is the configuration for the buoy server
	Config struct {
		Version    int                     `json:"version"`
		RootDomain string                  `json:"rootDomain"`
		Services   map[string]ServiceGroup `json:"services"`
		ConfigDir  string                  `json:"-"`
		ConfigFile string                  `json:"-"`
	}

	// ServiceGroup is a group of services that share the same domain
	ServiceGroup map[string]Service

	// Service is a service that is running on a port
	Service struct {
		Port int `json:"port"`
	}
)

const (
	// DefaultConfigFilename is the default path to the config file
	DefaultConfigFilename = "buoy.json"
)

var defaultConfig = Config{
	Version:    1,
	RootDomain: "b.com",
	Services: map[string]ServiceGroup{
		"example": {
			"/": Service{
				Port: 8080,
			},
		},
	},
}

// CreateConfig creates a default config file in the given directory
func CreateConfig(configDir string) error {
	configDir = getAbsPath(configDir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Printf("creating config directory %s...", configDir)
		err := os.Mkdir(configDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create config directory %s: %s", configDir, err)
		}
		fmt.Printf("done\n")
	} else {
		fmt.Printf("config directory already exists %s\n", configDir)
	}

	// TODO: Maybe have a force option to overwrite the config file if it exists
	configFilePath := path.Join(configDir, DefaultConfigFilename)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Printf("creating config file %s...", configFilePath)
		file, err := os.Create(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to create config file %s: %s", configFilePath, err)
		}

		defer file.Close()

		json, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %s", err)
		}

		_, err = file.Write(json)
		if err != nil {
			return fmt.Errorf("failed to write config file %s: %s", configFilePath, err)
		}

		fmt.Printf("done\n")
	} else {
		fmt.Printf("config file %s already exists\n", configFilePath)
	}

	return nil
}

// LoadConfig loads the config file from the given directory
func LoadConfig(configDir string) (Config, error) {
	configDir = getAbsPath(configDir)
	configFilePath := path.Join(configDir, DefaultConfigFilename)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config file %s does not exist", configFilePath)
	}

	file, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file %s: %s", configFilePath, err)
	}

	defer file.Close()

	config := Config{
		ConfigDir:  configDir,
		ConfigFile: configFilePath,
	}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to decode config file %s: %s", configFilePath, err)
	}

	return config, nil
}

// GetSSLCertPath returns the path to the SSL cert
func (c Config) GetSSLCertPath() string {
	return path.Join(c.ConfigDir, "cert.pem")
}

// GetSSLKeyPath returns the path to the SSL key
func (c Config) GetSSLKeyPath() string {
	return path.Join(c.ConfigDir, "key.pem")
}
