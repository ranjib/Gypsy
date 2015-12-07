// Copyright 2015 Ranjib Dey.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
