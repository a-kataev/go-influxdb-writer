package writer

import "log"

type Logger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Infof(template string, args ...interface{}) {
	log.Printf("INFO writer: "+template, args...)
}

func (l *defaultLogger) Errorf(template string, args ...interface{}) {
	log.Printf("ERROR writer: "+template, args...)
}
