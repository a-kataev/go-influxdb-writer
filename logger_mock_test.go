package writer

import "fmt"

type mockLogger struct {
	InfoLines  []string
	ErrorLines []string
}

func (l *mockLogger) Infof(template string, args ...interface{}) {
	l.InfoLines = append(l.InfoLines, fmt.Sprintf(template, args...))
}

func (l *mockLogger) Errorf(template string, args ...interface{}) {
	l.ErrorLines = append(l.ErrorLines, fmt.Sprintf(template, args...))
}
