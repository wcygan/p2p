package logger

import (
	"log/slog"
	"os"
	"strings"

	"example.com/p2p/pkg/config"
)

// Logger wraps slog.Logger with application-specific functionality
type Logger struct {
	*slog.Logger
	level  slog.Level
	format string
}

// New creates a new logger based on the provided configuration
func New(cfg *config.Config) *Logger {
	level := parseLevel(cfg.LogLevel)
	format := strings.ToLower(cfg.LogFormat)
	
	// Normalize format to valid values
	if format != "json" {
		format = "text"
	}
	
	var handler slog.Handler
	
	handlerOpts := &slog.HandlerOptions{
		Level: level,
		AddSource: level <= slog.LevelDebug, // Add source location for debug and lower
	}
	
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	default:
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}
	
	logger := &Logger{
		Logger: slog.New(handler),
		level:  level,
		format: format,
	}
	
	return logger
}

// parseLevel converts string log level to slog.Level
func parseLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// IsEnabled returns true if the given level is enabled
func (l *Logger) IsEnabled(level slog.Level) bool {
	return level >= l.level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() slog.Level {
	return l.level
}

// GetFormat returns the current log format
func (l *Logger) GetFormat() string {
	return l.format
}

// WithPeer returns a logger with peer-specific context
func (l *Logger) WithPeer(peerID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("peer_id", peerID),
		level:  l.level,
		format: l.format,
	}
}

// WithConnection returns a logger with connection-specific context
func (l *Logger) WithConnection(connID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("conn_id", connID),
		level:  l.level,
		format: l.format,
	}
}

// WithMessage returns a logger with message-specific context
func (l *Logger) WithMessage(msgType, senderID string, seqNo int) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			"msg_type", msgType,
			"sender_id", senderID,
			"seq_no", seqNo,
		),
		level:  l.level,
		format: l.format,
	}
}

// LogPeerConnected logs when a peer connects
func (l *Logger) LogPeerConnected(peerID, addr string) {
	l.Info("Peer connected",
		"peer_id", peerID,
		"address", addr,
		"event", "peer_connected",
	)
}

// LogPeerDisconnected logs when a peer disconnects
func (l *Logger) LogPeerDisconnected(peerID, reason string) {
	l.Info("Peer disconnected",
		"peer_id", peerID,
		"reason", reason,
		"event", "peer_disconnected",
	)
}

// LogMessageSent logs when a message is sent
func (l *Logger) LogMessageSent(msgType, recipientID string, seqNo int) {
	l.Debug("Message sent",
		"msg_type", msgType,
		"recipient_id", recipientID,
		"seq_no", seqNo,
		"event", "message_sent",
	)
}

// LogMessageReceived logs when a message is received
func (l *Logger) LogMessageReceived(msgType, senderID string, seqNo int) {
	l.Debug("Message received",
		"msg_type", msgType,
		"sender_id", senderID,
		"seq_no", seqNo,
		"event", "message_received",
	)
}

// LogMessageBroadcast logs when a message is broadcast
func (l *Logger) LogMessageBroadcast(msgType string, seqNo, peerCount int) {
	l.Debug("Message broadcast",
		"msg_type", msgType,
		"seq_no", seqNo,
		"peer_count", peerCount,
		"event", "message_broadcast",
	)
}

// LogHeartbeatSent logs when a heartbeat is sent
func (l *Logger) LogHeartbeatSent(peerID string, seqNo int) {
	l.Debug("Heartbeat sent",
		"peer_id", peerID,
		"seq_no", seqNo,
		"event", "heartbeat_sent",
	)
}

// LogHeartbeatReceived logs when a heartbeat is received
func (l *Logger) LogHeartbeatReceived(peerID string, seqNo int) {
	l.Debug("Heartbeat received",
		"peer_id", peerID,
		"seq_no", seqNo,
		"event", "heartbeat_received",
	)
}

// LogPeerTimedOut logs when a peer times out
func (l *Logger) LogPeerTimedOut(peerID string, lastSeen string) {
	l.Warn("Peer timed out",
		"peer_id", peerID,
		"last_seen", lastSeen,
		"event", "peer_timeout",
	)
}

// LogConnectionError logs connection errors
func (l *Logger) LogConnectionError(addr string, err error) {
	l.Error("Connection error",
		"address", addr,
		"error", err.Error(),
		"event", "connection_error",
	)
}

// LogConfigLoaded logs when configuration is loaded
func (l *Logger) LogConfigLoaded(source string, peerCount int) {
	l.Info("Configuration loaded",
		"source", source,
		"peer_count", peerCount,
		"event", "config_loaded",
	)
}

// LogServerStarted logs when the server starts
func (l *Logger) LogServerStarted(peerID, addr string) {
	l.Info("P2P server started",
		"peer_id", peerID,
		"listen_addr", addr,
		"event", "server_started",
	)
}

// LogServerStopped logs when the server stops
func (l *Logger) LogServerStopped(peerID string) {
	l.Info("P2P server stopped",
		"peer_id", peerID,
		"event", "server_stopped",
	)
}

// LogStatistics logs periodic statistics
func (l *Logger) LogStatistics(stats map[string]interface{}) {
	args := make([]interface{}, 0, len(stats)*2+2)
	args = append(args, "event", "statistics")
	
	for key, value := range stats {
		args = append(args, key, value)
	}
	
	l.Info("Statistics", args...)
}