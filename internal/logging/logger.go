package logging

import (
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	// AppName is the application name for logging.
	AppName = "jx-app-sonar-scanner"
)

var (
	logger = log.WithFields(log.Fields{"app": AppName})
)

func init() {
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
}

// AppLogger returns the application logger.
func AppLogger() *log.Entry {
	return logger
}

// SetLevel sets the logging level
func SetLevel(s string) error {
	level, err := log.ParseLevel(s)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}
