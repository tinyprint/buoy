package config

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "create_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpConfigDir := path.Join(tmpDir, ".buoy")

	err = CreateConfig(tmpConfigDir)
	if err != nil {
		t.Fatal(err)
	}

	// check that the config directory was created
	if _, err := os.Stat(tmpConfigDir); os.IsNotExist(err) {
		t.Errorf("config directory not created at %s", tmpConfigDir)
	}

	// check that the config file was created
	expectedConfigFile := filepath.Join(tmpConfigDir, DefaultConfigFilename)
	if _, err := os.Stat(expectedConfigFile); os.IsNotExist(err) {
		t.Errorf("config file not created at %s", expectedConfigFile)
	}

	// check that the config file is valid json
	jsonFile, err := os.Open(expectedConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var config map[string]interface{}
	err = decoder.Decode(&config)
	if err != nil {
		t.Fatal(err)
	}

	version := config["version"].(float64)

	if version != 1 {
		t.Errorf("expected version 1, got '%v'", version)
	}
}
