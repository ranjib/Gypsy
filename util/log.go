package util

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

func ConfigureLogging(level, format string, output io.Writer) error {
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		return errors.New("Invalid log level: " + level)
	}

	switch format {
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	default:
		return errors.New("Invalid format type: " + format)
	}
	log.SetOutput(output)
	return nil

}
