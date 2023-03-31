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

// GetConfigDir returns the config directory if it exists
func GetConfigDir(configDir string) (string, error) {
	configDir = getAbsPath(configDir)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return "", fmt.Errorf("config directory %s does not exist", configDir)
	}

	return configDir, nil
}

// SetupConfigDir creates the config directory if it doesn't exist
func SetupConfigDir(uid int, gid int, configDir string) (string, error) {
	configDir = getAbsPath(configDir)

	fmt.Printf("# creating config directory %s...", configDir)
	if _, err := os.Stat(configDir); os.IsExist(err) {
		fmt.Println("already exists")
		return configDir, nil
	}

	err := os.Mkdir(configDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %s", configDir, err)
	}

	chownError := os.Chown(configDir, uid, gid)
	if chownError != nil {
		return "", fmt.Errorf("error changing config directory owner: %s", chownError)
	}
	fmt.Println("done")

	return configDir, nil
}
