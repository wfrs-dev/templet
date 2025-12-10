package utils

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Routes map[string]string `json:"routes"`
	Debug  bool              `json:"debug"`
}

func ConfigFileName(filename string) string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	dir = filepath.Join(dir, "templet")
	os.MkdirAll(dir, 0755)

	return filepath.Join(dir, filename)
}

func ReadConfig() (Config, error) {
	var config Config

	filename := ConfigFileName("config.json")
	if _, err := os.Stat(filename); err != nil {
		return config, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// GetStdin obtiene la entroda por stdin o una cadena vacía si no hay
func GetStdin() string {
	stat, _ := os.Stdin.Stat()
	sb := strings.Builder{}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			sb.WriteString(scanner.Text() + " ")
		}
		if err := scanner.Err(); err != nil {
			return ""
		}
	}

	return sb.String()
}

// Exists checks if a file or directory exists
func Exists(path string, isDir ...bool) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if len(isDir) > 0 && isDir[0] {
		return info.IsDir()
	}

	return true
}
