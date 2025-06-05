package logger

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"example.com/p2p/pkg/config"
)

func TestNew(t *testing.T) {
	cfg := config.Default()
	cfg.LogLevel = "debug"
	cfg.LogFormat = "json"
	
	logger := New(cfg)
	
	if logger.GetLevel() != slog.LevelDebug {
		t.Errorf("expected debug level, got %v", logger.GetLevel())
	}
	
	if logger.GetFormat() != "json" {
		t.Errorf("expected json format, got %s", logger.GetFormat())
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // defaults to info
		{"", slog.LevelInfo},        // defaults to info
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsEnabled(t *testing.T) {
	cfg := config.Default()
	cfg.LogLevel = "warn"
	
	logger := New(cfg)
	
	if logger.IsEnabled(slog.LevelDebug) {
		t.Error("debug should not be enabled when level is warn")
	}
	
	if logger.IsEnabled(slog.LevelInfo) {
		t.Error("info should not be enabled when level is warn")
	}
	
	if !logger.IsEnabled(slog.LevelWarn) {
		t.Error("warn should be enabled when level is warn")
	}
	
	if !logger.IsEnabled(slog.LevelError) {
		t.Error("error should be enabled when level is warn")
	}
}

func TestWithPeer(t *testing.T) {
	cfg := config.Default()
	logger := New(cfg)
	
	peerLogger := logger.WithPeer("peer123")
	
	if peerLogger.GetLevel() != logger.GetLevel() {
		t.Error("peer logger should inherit log level")
	}
	
	if peerLogger.GetFormat() != logger.GetFormat() {
		t.Error("peer logger should inherit log format")
	}
}

func TestWithConnection(t *testing.T) {
	cfg := config.Default()
	logger := New(cfg)
	
	connLogger := logger.WithConnection("conn456")
	
	if connLogger.GetLevel() != logger.GetLevel() {
		t.Error("connection logger should inherit log level")
	}
}

func TestWithMessage(t *testing.T) {
	cfg := config.Default()
	logger := New(cfg)
	
	msgLogger := logger.WithMessage("chat", "peer123", 42)
	
	if msgLogger.GetLevel() != logger.GetLevel() {
		t.Error("message logger should inherit log level")
	}
}

func TestTextFormat(t *testing.T) {
	// Create logger with text format
	cfg := config.Default()
	cfg.LogLevel = "info"
	cfg.LogFormat = "text"
	
	// Temporarily redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	logger := New(cfg)
	logger.Info("test message", "key", "value")
	
	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	
	var output [1024]byte
	n, _ := r.Read(output[:])
	result := string(output[:n])
	
	if !strings.Contains(result, "test message") {
		t.Errorf("expected log to contain 'test message', got: %s", result)
	}
	
	if !strings.Contains(result, "key=value") {
		t.Errorf("expected log to contain 'key=value', got: %s", result)
	}
}

func TestJSONFormat(t *testing.T) {
	// Create logger with JSON format
	cfg := config.Default()
	cfg.LogLevel = "info"
	cfg.LogFormat = "json"
	
	// Temporarily redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	logger := New(cfg)
	logger.Info("test message", "key", "value")
	
	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	
	var output [1024]byte
	n, _ := r.Read(output[:])
	result := string(output[:n])
	
	// Parse as JSON to verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(result), &logEntry); err != nil {
		t.Errorf("log output is not valid JSON: %v\nOutput: %s", err, result)
	}
	
	if logEntry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", logEntry["msg"])
	}
	
	if logEntry["key"] != "value" {
		t.Errorf("expected key 'value', got %v", logEntry["key"])
	}
}

func TestSpecificLogMethods(t *testing.T) {
	cfg := config.Default()
	cfg.LogLevel = "debug"
	
	// We can't easily test the output without complex stdout capture,
	// but we can test that the methods don't panic
	logger := New(cfg)
	
	// Test all specific log methods
	logger.LogPeerConnected("peer123", "localhost:8080")
	logger.LogPeerDisconnected("peer123", "timeout")
	logger.LogMessageSent("chat", "peer456", 1)
	logger.LogMessageReceived("chat", "peer456", 1)
	logger.LogMessageBroadcast("chat", 1, 5)
	logger.LogHeartbeatSent("peer123", 1)
	logger.LogHeartbeatReceived("peer123", 1)
	logger.LogPeerTimedOut("peer123", "2023-01-01T00:00:00Z")
	logger.LogConnectionError("localhost:8080", &testError{"connection failed"})
	logger.LogConfigLoaded("file", 3)
	logger.LogServerStarted("peer123", "localhost:8080")
	logger.LogServerStopped("peer123")
	
	stats := map[string]interface{}{
		"connections": 5,
		"messages":    100,
		"uptime":      "1h30m",
	}
	logger.LogStatistics(stats)
}

func TestLoggerWithDifferentLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := config.Default()
			cfg.LogLevel = level
			
			logger := New(cfg)
			expectedLevel := parseLevel(level)
			
			if logger.GetLevel() != expectedLevel {
				t.Errorf("expected level %v, got %v", expectedLevel, logger.GetLevel())
			}
		})
	}
}

func TestLoggerFormats(t *testing.T) {
	formats := []string{"text", "json", "TEXT", "JSON", "invalid"}
	
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cfg := config.Default()
			cfg.LogFormat = format
			
			logger := New(cfg)
			
			expectedFormat := strings.ToLower(format)
			if expectedFormat != "text" && expectedFormat != "json" {
				expectedFormat = "text" // default
			}
			
			if logger.GetFormat() != expectedFormat {
				t.Errorf("expected format %s, got %s", expectedFormat, logger.GetFormat())
			}
		})
	}
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestLoggerChaining(t *testing.T) {
	cfg := config.Default()
	logger := New(cfg)
	
	// Test that we can chain context methods
	chainedLogger := logger.WithPeer("peer123").WithConnection("conn456")
	
	if chainedLogger.GetLevel() != logger.GetLevel() {
		t.Error("chained logger should inherit log level")
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	cfg := config.Default()
	cfg.LogLevel = "info"
	logger := New(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}

func BenchmarkLoggerDebugDisabled(b *testing.B) {
	cfg := config.Default()
	cfg.LogLevel = "info" // debug is disabled
	logger := New(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message", "iteration", i)
	}
}

func BenchmarkLoggerWithContext(b *testing.B) {
	cfg := config.Default()
	cfg.LogLevel = "info"
	logger := New(cfg).WithPeer("peer123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}