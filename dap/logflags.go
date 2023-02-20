package dap

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"sort"
	"time"
)

//var dap = false

// DAPLogger returns a logger for dap package.
func DAPLogger() *logrus.Entry {
	return makeLogger(false, logrus.Fields{"layer": "dap"})
}

var logOut io.WriteCloser

func makeLogger(flag bool, fields logrus.Fields) *logrus.Entry {
	logger := logrus.New().WithFields(fields)
	logger.Logger.Formatter = &textFormatter{}
	if logOut != nil {
		logger.Logger.Out = logOut
	}
	logger.Logger.Level = logrus.DebugLevel
	if !flag {
		logger.Logger.Level = logrus.InfoLevel
	}
	return logger
}

// textFormatter is a simplified version of logrus.TextFormatter that
// doesn't make logs unreadable when they are output to a text file or to a
// terminal that doesn't support colors.
type textFormatter struct {
}

func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteString(entry.Time.Format(time.RFC3339))
	b.WriteByte(' ')
	b.WriteString(entry.Level.String())
	b.WriteByte(' ')
	for i, key := range keys {
		b.WriteString(key)
		b.WriteByte('=')
		stringVal, ok := entry.Data[key].(string)
		if !ok {
			stringVal = fmt.Sprint(entry.Data[key])
		}
		if f.needsQuoting(stringVal) {
			fmt.Fprintf(b, "%q", stringVal)
		} else {
			b.WriteString(stringVal)
		}
		if i != len(keys)-1 {
			b.WriteByte(',')
		} else {
			b.WriteByte(' ')
		}
	}
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *textFormatter) needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}
