package buoy

import (
	"fmt"
	"os"
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

// GetOrCreateConfigDir creates the config directory if it doesn't exist
func GetOrCreateConfigDir(configDir string) (string, error) {
	configDir = getAbsPath(configDir)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Printf("# creating config directory %s...", configDir)
		err := os.Mkdir(configDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create config directory %s: %s", configDir, err)
		}
		fmt.Println("done")
	} else {
		fmt.Printf("# config directory already exists %s\n", configDir)
	}

	return configDir, nil
}
