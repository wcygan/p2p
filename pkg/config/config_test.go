package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	config := Default()
	
	if config.ListenAddr != "localhost:0" {
		t.Errorf("expected default listen_addr 'localhost:0', got %s", config.ListenAddr)
	}
	
	if len(config.Peers) != 0 {
		t.Errorf("expected empty peers list, got %v", config.Peers)
	}
	
	if config.MaxConnections != 50 {
		t.Errorf("expected default max_connections 50, got %d", config.MaxConnections)
	}
	
	if time.Duration(config.HeartbeatInterval) != 30*time.Second {
		t.Errorf("expected default heartbeat_interval 30s, got %v", config.HeartbeatInterval)
	}
	
	if config.LogLevel != "info" {
		t.Errorf("expected default log_level 'info', got %s", config.LogLevel)
	}
}

func TestLoadFromFileJSON(t *testing.T) {
	// Create temporary JSON config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	
	jsonConfig := `{
		"listen_addr": "localhost:8080",
		"peers": ["localhost:8081", "localhost:8082"],
		"max_connections": 25,
		"connect_timeout": "5s",
		"heartbeat_interval": "15s",
		"log_level": "debug"
	}`
	
	if err := os.WriteFile(configFile, []byte(jsonConfig), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	
	config, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	
	if config.ListenAddr != "localhost:8080" {
		t.Errorf("expected listen_addr 'localhost:8080', got %s", config.ListenAddr)
	}
	
	expectedPeers := []string{"localhost:8081", "localhost:8082"}
	if !reflect.DeepEqual(config.Peers, expectedPeers) {
		t.Errorf("expected peers %v, got %v", expectedPeers, config.Peers)
	}
	
	if config.MaxConnections != 25 {
		t.Errorf("expected max_connections 25, got %d", config.MaxConnections)
	}
	
	if time.Duration(config.ConnectTimeout) != 5*time.Second {
		t.Errorf("expected connect_timeout 5s, got %v", config.ConnectTimeout)
	}
	
	if config.LogLevel != "debug" {
		t.Errorf("expected log_level 'debug', got %s", config.LogLevel)
	}
}

func TestLoadFromFileYAML(t *testing.T) {
	// Create temporary YAML config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	
	yamlConfig := `listen_addr: "localhost:9090"
peers: ["localhost:9091", "localhost:9092"]
max_connections: 30
connect_timeout: "8s"
heartbeat_interval: "20s"
log_level: "warn"
log_format: "json"`
	
	if err := os.WriteFile(configFile, []byte(yamlConfig), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	
	config, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	
	if config.ListenAddr != "localhost:9090" {
		t.Errorf("expected listen_addr 'localhost:9090', got %s", config.ListenAddr)
	}
	
	expectedPeers := []string{"localhost:9091", "localhost:9092"}
	if !reflect.DeepEqual(config.Peers, expectedPeers) {
		t.Errorf("expected peers %v, got %v", expectedPeers, config.Peers)
	}
	
	if config.MaxConnections != 30 {
		t.Errorf("expected max_connections 30, got %d", config.MaxConnections)
	}
	
	if config.LogFormat != "json" {
		t.Errorf("expected log_format 'json', got %s", config.LogFormat)
	}
}

func TestLoadFromFileInvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.txt")
	
	if err := os.WriteFile(configFile, []byte("invalid"), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	
	_, err := LoadFromFile(configFile)
	if err == nil {
		t.Fatal("expected error for unsupported file format")
	}
}

func TestLoadFromFileNonExistent(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/config.json")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestLoadFromFileInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	
	if err := os.WriteFile(configFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	
	_, err := LoadFromFile(configFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original environment
	origVars := map[string]string{
		"P2P_LISTEN_ADDR":         os.Getenv("P2P_LISTEN_ADDR"),
		"P2P_PEERS":               os.Getenv("P2P_PEERS"),
		"P2P_MAX_CONNECTIONS":     os.Getenv("P2P_MAX_CONNECTIONS"),
		"P2P_CONNECT_TIMEOUT":     os.Getenv("P2P_CONNECT_TIMEOUT"),
		"P2P_HEARTBEAT_INTERVAL":  os.Getenv("P2P_HEARTBEAT_INTERVAL"),
		"P2P_LOG_LEVEL":           os.Getenv("P2P_LOG_LEVEL"),
	}
	
	// Restore environment after test
	defer func() {
		for key, value := range origVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	// Set test environment variables
	os.Setenv("P2P_LISTEN_ADDR", "localhost:7070")
	os.Setenv("P2P_PEERS", "localhost:7071,localhost:7072,localhost:7073")
	os.Setenv("P2P_MAX_CONNECTIONS", "40")
	os.Setenv("P2P_CONNECT_TIMEOUT", "12s")
	os.Setenv("P2P_HEARTBEAT_INTERVAL", "25s")
	os.Setenv("P2P_LOG_LEVEL", "error")
	
	config := LoadFromEnv()
	
	if config.ListenAddr != "localhost:7070" {
		t.Errorf("expected listen_addr 'localhost:7070', got %s", config.ListenAddr)
	}
	
	expectedPeers := []string{"localhost:7071", "localhost:7072", "localhost:7073"}
	if !reflect.DeepEqual(config.Peers, expectedPeers) {
		t.Errorf("expected peers %v, got %v", expectedPeers, config.Peers)
	}
	
	if config.MaxConnections != 40 {
		t.Errorf("expected max_connections 40, got %d", config.MaxConnections)
	}
	
	if time.Duration(config.ConnectTimeout) != 12*time.Second {
		t.Errorf("expected connect_timeout 12s, got %v", config.ConnectTimeout)
	}
	
	if time.Duration(config.HeartbeatInterval) != 25*time.Second {
		t.Errorf("expected heartbeat_interval 25s, got %v", config.HeartbeatInterval)
	}
	
	if config.LogLevel != "error" {
		t.Errorf("expected log_level 'error', got %s", config.LogLevel)
	}
}

func TestLoadFromEnvDefaults(t *testing.T) {
	// Clear all P2P environment variables
	envVars := []string{
		"P2P_LISTEN_ADDR", "P2P_PEERS", "P2P_MAX_CONNECTIONS",
		"P2P_CONNECT_TIMEOUT", "P2P_HEARTBEAT_INTERVAL", "P2P_LOG_LEVEL",
	}
	
	origValues := make(map[string]string)
	for _, env := range envVars {
		origValues[env] = os.Getenv(env)
		os.Unsetenv(env)
	}
	
	defer func() {
		for env, value := range origValues {
			if value != "" {
				os.Setenv(env, value)
			}
		}
	}()
	
	config := LoadFromEnv()
	defaultConfig := Default()
	
	if !reflect.DeepEqual(config, defaultConfig) {
		t.Errorf("expected env config to equal defaults when no env vars set")
	}
}

func TestLoad(t *testing.T) {
	// Test loading with no config file
	config, err := Load("")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	
	if config.ListenAddr != "localhost:0" {
		t.Errorf("expected default listen_addr, got %s", config.ListenAddr)
	}
	
	// Test loading with config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	
	jsonConfig := `{"listen_addr": "localhost:6060"}`
	if err := os.WriteFile(configFile, []byte(jsonConfig), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	
	config, err = Load(configFile)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	
	if config.ListenAddr != "localhost:6060" {
		t.Errorf("expected listen_addr 'localhost:6060', got %s", config.ListenAddr)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Default(),
			wantErr: false,
		},
		{
			name: "empty listen addr",
			config: &Config{
				ListenAddr:        "",
				MaxConnections:    1,
				ConnectTimeout:    JSONDuration(time.Second),
				HeartbeatInterval: JSONDuration(time.Second),
				HeartbeatTimeout:  JSONDuration(time.Second),
				MessageBufferSize: 1,
				DedupCacheSize:    1,
				LogLevel:          "info",
				LogFormat:         "text",
			},
			wantErr: true,
		},
		{
			name: "zero max connections",
			config: &Config{
				ListenAddr:        "localhost:0",
				MaxConnections:    0,
				ConnectTimeout:    JSONDuration(time.Second),
				HeartbeatInterval: JSONDuration(time.Second),
				HeartbeatTimeout:  JSONDuration(time.Second),
				MessageBufferSize: 1,
				DedupCacheSize:    1,
				LogLevel:          "info",
				LogFormat:         "text",
			},
			wantErr: true,
		},
		{
			name: "zero connect timeout",
			config: &Config{
				ListenAddr:        "localhost:0",
				MaxConnections:    1,
				ConnectTimeout:    JSONDuration(0),
				HeartbeatInterval: JSONDuration(time.Second),
				HeartbeatTimeout:  JSONDuration(time.Second),
				MessageBufferSize: 1,
				DedupCacheSize:    1,
				LogLevel:          "info",
				LogFormat:         "text",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				ListenAddr:        "localhost:0",
				MaxConnections:    1,
				ConnectTimeout:    JSONDuration(time.Second),
				HeartbeatInterval: JSONDuration(time.Second),
				HeartbeatTimeout:  JSONDuration(time.Second),
				MessageBufferSize: 1,
				DedupCacheSize:    1,
				LogLevel:          "invalid",
				LogFormat:         "text",
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			config: &Config{
				ListenAddr:        "localhost:0",
				MaxConnections:    1,
				ConnectTimeout:    JSONDuration(time.Second),
				HeartbeatInterval: JSONDuration(time.Second),
				HeartbeatTimeout:  JSONDuration(time.Second),
				MessageBufferSize: 1,
				DedupCacheSize:    1,
				LogLevel:          "info",
				LogFormat:         "invalid",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveToFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "save_test.json")
	
	config := Default()
	config.ListenAddr = "localhost:5050"
	config.LogLevel = "debug"
	
	if err := config.SaveToFile(configFile); err != nil {
		t.Fatalf("save config: %v", err)
	}
	
	// Load back and verify
	loaded, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("load saved config: %v", err)
	}
	
	if loaded.ListenAddr != config.ListenAddr {
		t.Errorf("expected listen_addr %s, got %s", config.ListenAddr, loaded.ListenAddr)
	}
	
	if loaded.LogLevel != config.LogLevel {
		t.Errorf("expected log_level %s, got %s", config.LogLevel, loaded.LogLevel)
	}
}

func TestParseYAMLEdgeCases(t *testing.T) {
	config := Default()
	
	yamlData := `# This is a comment
listen_addr: "localhost:1234"
# Another comment
peers: []
max_connections: 100

log_level: debug`
	
	if err := parseYAML([]byte(yamlData), config); err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	
	if config.ListenAddr != "localhost:1234" {
		t.Errorf("expected listen_addr 'localhost:1234', got %s", config.ListenAddr)
	}
	
	if config.MaxConnections != 100 {
		t.Errorf("expected max_connections 100, got %d", config.MaxConnections)
	}
	
	if config.LogLevel != "debug" {
		t.Errorf("expected log_level 'debug', got %s", config.LogLevel)
	}
}