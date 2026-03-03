package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/natefinch/lumberjack.v2"
)

var std *log.Logger

func Init(logDir string) error {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log dir: %w", err)
	}

	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "moralink-gost.log"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	multi := io.MultiWriter(os.Stdout, logFile)
	std = log.New(multi, "", log.LstdFlags|log.Lshortfile)
	log.SetOutput(multi)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Switching std place", logDir)
	return nil
}

func logPath() string {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("PROGRAMDATA")
		if base == "" {
			base = `C:\ProgramData`
		}
		return filepath.Join(base, "moralink-gost", "logs")
	default:
		return "/var/log/moralink-gost"
	}
}

// InitDefault initialises logging to the OS-appropriate default path.
func InitDefault() {
	if err := Init(logPath()); err != nil {
		// last resort: at least log to stdout
		log.Printf("⚠️  Could not init file logger: %v", err)
	}
}

// Convenience wrappers so other packages just import logger and call logger.Info(...)
func Info(v ...any)             { std.SetPrefix("[INFO]  "); std.Println(v...) }
func Warn(v ...any)             { std.SetPrefix("[WARN]  "); std.Println(v...) }
func Error(v ...any)            { std.SetPrefix("[ERROR] "); std.Println(v...) }
func Infof(f string, v ...any)  { std.SetPrefix("[INFO]  "); std.Printf(f+"\n", v...) }
func Warnf(f string, v ...any)  { std.SetPrefix("[WARN]  "); std.Printf(f+"\n", v...) }
func Errorf(f string, v ...any) { std.SetPrefix("[ERROR] "); std.Printf(f+"\n", v...) }
