// internal/util/helpers.go
package util

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port                int           `yaml:"port"`
	ReadTimeout         time.Duration `yaml:"read_timeout"`
	WriteTimeout        time.Duration `yaml:"write_timeout"`
	WebSocketPingInterval time.Duration `yaml:"websocket_ping_interval"`
	MaxRooms            int           `yaml:"max_rooms"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Convert seconds to durations
	cfg.Server.ReadTimeout *= time.Second
	cfg.Server.WriteTimeout *= time.Second
	cfg.Server.WebSocketPingInterval *= time.Second

	return &cfg, nil
}


func JSONEncode(data interface{}) ([]byte, error) {
    return json.Marshal(data)
}


func JSONDecode(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}


func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}


func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(data)
}