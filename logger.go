package writer

import "log"

// Logger
type Logger interface {
	// Infof
	Infof(template string, args ...interface{})
	// Errorf
	Errorf(template string, args ...interface{})
}

type defaultLogger struct{}

// Infof
func (l *defaultLogger) Infof(template string, args ...interface{}) {
	log.Printf("INFO writer: "+template, args...)
}

// Errorf
func (l *defaultLogger) Errorf(template string, args ...interface{}) {
	log.Printf("ERROR writer: "+template, args...)
}
