package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Network settings
	ListenAddr string   `json:"listen_addr" yaml:"listen_addr"`
	Peers      []string `json:"peers" yaml:"peers"`
	
	// Connection settings
	MaxConnections    int              `json:"max_connections" yaml:"max_connections"`
	ConnectTimeout    JSONDuration     `json:"connect_timeout" yaml:"connect_timeout"`
	HeartbeatInterval JSONDuration     `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	HeartbeatTimeout  JSONDuration     `json:"heartbeat_timeout" yaml:"heartbeat_timeout"`
	
	// Message settings
	MessageBufferSize int `json:"message_buffer_size" yaml:"message_buffer_size"`
	DedupCacheSize    int `json:"dedup_cache_size" yaml:"dedup_cache_size"`
	
	// Logging settings
	LogLevel  string `json:"log_level" yaml:"log_level"`
	LogFormat string `json:"log_format" yaml:"log_format"` // "json" or "text"
}

// JSONDuration wraps time.Duration to provide JSON marshaling/unmarshaling
type JSONDuration time.Duration

func (d JSONDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *JSONDuration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = JSONDuration(dur)
	return nil
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		ListenAddr:        "localhost:0",
		Peers:             []string{},
		MaxConnections:    50,
		ConnectTimeout:    JSONDuration(10 * time.Second),
		HeartbeatInterval: JSONDuration(30 * time.Second),
		HeartbeatTimeout:  JSONDuration(5 * time.Second),
		MessageBufferSize: 16,
		DedupCacheSize:    100,
		LogLevel:          "info",
		LogFormat:         "text",
	}
}

// LoadFromFile loads configuration from a JSON or YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	
	config := Default()
	
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := parseYAML(data, config); err != nil {
			return nil, fmt.Errorf("parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}
	
	return config, nil
}

// LoadFromEnv loads configuration from environment variables
// Environment variables use P2P_ prefix and underscore notation
func LoadFromEnv() *Config {
	config := Default()
	
	if addr := os.Getenv("P2P_LISTEN_ADDR"); addr != "" {
		config.ListenAddr = addr
	}
	
	if peers := os.Getenv("P2P_PEERS"); peers != "" {
		config.Peers = strings.Split(peers, ",")
		// Trim whitespace from each peer
		for i, peer := range config.Peers {
			config.Peers[i] = strings.TrimSpace(peer)
		}
	}
	
	if maxConn := os.Getenv("P2P_MAX_CONNECTIONS"); maxConn != "" {
		if val, err := strconv.Atoi(maxConn); err == nil {
			config.MaxConnections = val
		}
	}
	
	if timeout := os.Getenv("P2P_CONNECT_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			config.ConnectTimeout = JSONDuration(val)
		}
	}
	
	if interval := os.Getenv("P2P_HEARTBEAT_INTERVAL"); interval != "" {
		if val, err := time.ParseDuration(interval); err == nil {
			config.HeartbeatInterval = JSONDuration(val)
		}
	}
	
	if timeout := os.Getenv("P2P_HEARTBEAT_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			config.HeartbeatTimeout = JSONDuration(val)
		}
	}
	
	if bufSize := os.Getenv("P2P_MESSAGE_BUFFER_SIZE"); bufSize != "" {
		if val, err := strconv.Atoi(bufSize); err == nil {
			config.MessageBufferSize = val
		}
	}
	
	if cacheSize := os.Getenv("P2P_DEDUP_CACHE_SIZE"); cacheSize != "" {
		if val, err := strconv.Atoi(cacheSize); err == nil {
			config.DedupCacheSize = val
		}
	}
	
	if level := os.Getenv("P2P_LOG_LEVEL"); level != "" {
		config.LogLevel = level
	}
	
	if format := os.Getenv("P2P_LOG_FORMAT"); format != "" {
		config.LogFormat = format
	}
	
	return config
}

// Load loads configuration with the following precedence:
// 1. Config file (if provided)
// 2. Environment variables
// 3. Defaults
func Load(configFile string) (*Config, error) {
	var config *Config
	var err error
	
	if configFile != "" {
		config, err = LoadFromFile(configFile)
		if err != nil {
			return nil, err
		}
	} else {
		config = Default()
	}
	
	// Override with environment variables
	envConfig := LoadFromEnv()
	mergeConfigs(config, envConfig)
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen_addr cannot be empty")
	}
	
	if c.MaxConnections < 1 {
		return fmt.Errorf("max_connections must be at least 1")
	}
	
	if time.Duration(c.ConnectTimeout) <= 0 {
		return fmt.Errorf("connect_timeout must be positive")
	}
	
	if time.Duration(c.HeartbeatInterval) <= 0 {
		return fmt.Errorf("heartbeat_interval must be positive")
	}
	
	if time.Duration(c.HeartbeatTimeout) <= 0 {
		return fmt.Errorf("heartbeat_timeout must be positive")
	}
	
	if c.MessageBufferSize < 1 {
		return fmt.Errorf("message_buffer_size must be at least 1")
	}
	
	if c.DedupCacheSize < 1 {
		return fmt.Errorf("dedup_cache_size must be at least 1")
	}
	
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("invalid log_level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}
	
	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[strings.ToLower(c.LogFormat)] {
		return fmt.Errorf("invalid log_format: %s (must be json or text)", c.LogFormat)
	}
	
	return nil
}

// SaveToFile saves the configuration to a JSON file
func (c *Config) SaveToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	
	return nil
}

// mergeConfigs merges environment config into base config for non-zero values
func mergeConfigs(base, env *Config) {
	if env.ListenAddr != Default().ListenAddr {
		base.ListenAddr = env.ListenAddr
	}
	if len(env.Peers) > 0 {
		base.Peers = env.Peers
	}
	if env.MaxConnections != Default().MaxConnections {
		base.MaxConnections = env.MaxConnections
	}
	if env.ConnectTimeout != Default().ConnectTimeout {
		base.ConnectTimeout = env.ConnectTimeout
	}
	if env.HeartbeatInterval != Default().HeartbeatInterval {
		base.HeartbeatInterval = env.HeartbeatInterval
	}
	if env.HeartbeatTimeout != Default().HeartbeatTimeout {
		base.HeartbeatTimeout = env.HeartbeatTimeout
	}
	if env.MessageBufferSize != Default().MessageBufferSize {
		base.MessageBufferSize = env.MessageBufferSize
	}
	if env.DedupCacheSize != Default().DedupCacheSize {
		base.DedupCacheSize = env.DedupCacheSize
	}
	if env.LogLevel != Default().LogLevel {
		base.LogLevel = env.LogLevel
	}
	if env.LogFormat != Default().LogFormat {
		base.LogFormat = env.LogFormat
	}
}

// parseYAML is a simple YAML parser for our config struct
// This avoids adding a YAML dependency for this basic use case
func parseYAML(data []byte, config *Config) error {
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
		
		switch key {
		case "listen_addr":
			config.ListenAddr = value
		case "peers":
			if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				value = value[1 : len(value)-1]
				if value != "" {
					peers := strings.Split(value, ",")
					config.Peers = make([]string, len(peers))
					for i, peer := range peers {
						peer = strings.TrimSpace(peer)
						// Remove surrounding quotes
						if (strings.HasPrefix(peer, "\"") && strings.HasSuffix(peer, "\"")) ||
							(strings.HasPrefix(peer, "'") && strings.HasSuffix(peer, "'")) {
							peer = peer[1 : len(peer)-1]
						}
						config.Peers[i] = peer
					}
				}
			}
		case "max_connections":
			if val, err := strconv.Atoi(value); err == nil {
				config.MaxConnections = val
			}
		case "connect_timeout":
			if val, err := time.ParseDuration(value); err == nil {
				config.ConnectTimeout = JSONDuration(val)
			}
		case "heartbeat_interval":
			if val, err := time.ParseDuration(value); err == nil {
				config.HeartbeatInterval = JSONDuration(val)
			}
		case "heartbeat_timeout":
			if val, err := time.ParseDuration(value); err == nil {
				config.HeartbeatTimeout = JSONDuration(val)
			}
		case "message_buffer_size":
			if val, err := strconv.Atoi(value); err == nil {
				config.MessageBufferSize = val
			}
		case "dedup_cache_size":
			if val, err := strconv.Atoi(value); err == nil {
				config.DedupCacheSize = val
			}
		case "log_level":
			config.LogLevel = value
		case "log_format":
			config.LogFormat = value
		}
	}
	
	return nil
}