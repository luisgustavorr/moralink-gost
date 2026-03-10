package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var std *log.Logger

func Init(baseDir string) {
	logDir := filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		// Can't create dir — nothing we can do, fail silently
		return
	}

	roller := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "moralink-gost.log"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	var w io.Writer

	if service.Interactive() {
		w = io.MultiWriter(os.Stdout, roller)
	} else {
		w = roller
	}

	log.SetOutput(w)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("")

	std = log.New(w, "", log.LstdFlags|log.Lshortfile)

	log.Printf("✅ 📝 Logger initialized: %s", filepath.Join(logDir, "moralink-gost.log"))
}

func Info(v ...any) { std.SetPrefix("[INFO]  "); std.Println(v...) }
func Debug(v ...any) {
	if viper.Get("api.mode") == "dev" {
		std.SetPrefix("[DEBUG]  ")
		std.Println(v...)
	}

}
func Debugf(f string, v ...any) {
	if viper.Get("api.mode") == "dev" {
		std.SetPrefix("[DEBUG]  ")
		std.Printf(f+"\n", v...)
	}

}
func Warn(v ...any)             { std.SetPrefix("[WARN]  "); std.Println(v...) }
func Error(v ...any)            { std.SetPrefix("[ERROR] "); std.Println(v...) }
func Infof(f string, v ...any)  { std.SetPrefix("[INFO]  "); std.Printf(f+"\n", v...) }
func Warnf(f string, v ...any)  { std.SetPrefix("[WARN]  "); std.Printf(f+"\n", v...) }
func Errorf(f string, v ...any) { std.SetPrefix("[ERROR] "); std.Printf(f+"\n", v...) }
