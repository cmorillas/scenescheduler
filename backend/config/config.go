// backend/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// defaultConfigPath defines the fixed, non-overridable path to the configuration file.
const defaultConfigPath = "config.json"

// Config is the top-level struct that holds the entire application configuration.
type Config struct {
	MediaSource MediaSourceConfig `json:"mediaSource"`
	WebServer   WebServerConfig   `json:"webServer"`
	OBS         OBSConfig         `json:"obs"`
	Paths       PathsConfig       `json:"paths"`
	Scheduler   SchedulerConfig   `json:"scheduler"` // Section for scheduler-specific settings.
}

type MediaSourceConfig struct {
    VideoDeviceIdentifier string `json:"videoDeviceIdentifier"`
    AudioDeviceIdentifier string `json:"audioDeviceIdentifier"`
    Quality               string `json:"quality"`
    
    // Derived fields (populated at runtime by Manager, not from JSON)
    VideoWidth      int `json:"-"`
    VideoHeight     int `json:"-"`
    VideoBitrate    int `json:"-"`
    AudioChannels   int `json:"-"`
    AudioSampleRate int `json:"-"`
}

type WebServerConfig struct {
	Port            string `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	HlsPath         string `json:"hlsPath"`
	EnableTLS       bool   `json:"enableTls"`
	CertFilePath    string `json:"certFilePath"`
	KeyFilePath     string `json:"keyFilePath"`
	ReadTimeout     time.Duration `json:"-"`
	WriteTimeout    time.Duration `json:"-"`
	ShutdownTimeout time.Duration `json:"-"`
}

type OBSConfig struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Password          string `json:"password"`
	ReconnectInterval int    `json:"reconnectInterval"`
	ScheduleScene     string `json:"scheduleScene"`
	ScheduleSceneAux  string `json:"scheduleSceneAux"`
	SourceNamePrefix  string `json:"sourceNamePrefix"`
}

type PathsConfig struct {
	LogFile  string `json:"logFile"`
	Schedule string `json:"schedule"`
}

// SchedulerConfig holds settings specific to the scheduler module.
type SchedulerConfig struct {
	// DefaultSource is an optional source that will be activated when no other
	// program is scheduled to be active.
	DefaultSource DefaultSource `json:"defaultSource"`
}

// DefaultSource defines a backup source to be used by the scheduler.
type DefaultSource struct {
	Name          string      `json:"name"`
	InputKind     string      `json:"inputKind"`
	URI           string      `json:"uri"`
	InputSettings interface{} `json:"inputSettings"`
	Transform	  interface{} `json:"transform"`
}

// NewConfig loads, decodes, and validates the configuration from the fixed default path.
func NewConfig() (*Config, error) {
	data, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", defaultConfigPath, err)
	}

	cfg := &Config{}
	cfg.applyDefaults()

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", defaultConfigPath, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Printf("[CONFIG] Configuration loaded successfully from '%s'.\n", defaultConfigPath)
	return cfg, nil
}

func (c *Config) applyDefaults() {
	c.MediaSource.Quality = "low"
	c.WebServer.Port = "8080"
	c.WebServer.HlsPath = "hls"
	c.WebServer.ReadTimeout = 30 * time.Second
	c.WebServer.WriteTimeout = 30 * time.Second
	c.WebServer.ShutdownTimeout = 15 * time.Second
	c.OBS.Host = "localhost"
	c.OBS.Port = 4455
	c.OBS.ReconnectInterval = 15
	c.OBS.SourceNamePrefix = "_sched_"
	c.Paths.Schedule = "schedule.json"
}

func (c *Config) validate() error {
	if c.OBS.ScheduleScene == "" || c.OBS.ScheduleSceneAux == "" {
		return fmt.Errorf("obs.scheduleScene and obs.scheduleSceneAux are required fields")
	}
	if c.OBS.Password == "" {
		log.Println("[CONFIG] WARN: obs.password is empty. Ensure OBS WebSocket auth is disabled if this is intentional.")
	}
	if c.WebServer.User == "" || c.WebServer.Password == "" {
		log.Println("[CONFIG] WARN: webServer.user or webServer.password is empty. Web server authentication will be disabled.")
	}
	if c.WebServer.EnableTLS && (c.WebServer.CertFilePath == "" || c.WebServer.KeyFilePath == "") {
		return fmt.Errorf("webServer.certFilePath and webServer.keyFilePath are required when TLS is enabled")
	}

	// Validate hlsPath is a safe relative path
	if err := validateSafeRelativePath(c.WebServer.HlsPath, "webServer.hlsPath"); err != nil {
		return err
	}

	return nil
}

// validateSafeRelativePath ensures a path is relative, doesn't contain path traversal,
// and doesn't escape the current directory.
func validateSafeRelativePath(path, fieldName string) error {
	if path == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Reject absolute paths (Unix and Windows)
	if filepath.IsAbs(path) {
		return fmt.Errorf("%s must be a relative path, got absolute path: %s", fieldName, path)
	}

	// Reject paths starting with ../
	if strings.HasPrefix(path, "..") {
		return fmt.Errorf("%s cannot start with '..': %s", fieldName, path)
	}

	// Clean the path and check it doesn't escape current directory
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "/../") {
		return fmt.Errorf("%s contains directory traversal: %s", fieldName, path)
	}

	return nil
}

