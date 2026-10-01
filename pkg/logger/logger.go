package logger

import (
	"github.com/sirupsen/logrus"
)

// New builds a configured logrus logger.
//   - JSON output in prod (machine-parseable for log aggregators),
//     human-friendly text otherwise.
//   - level parsed from a string; defaults to info on bad input.
//   - service name attached to every line as a field.
func New(service, level string, prod bool) *logrus.Entry {
	l := logrus.New()

	if prod {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		l.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	l.SetLevel(lvl)

	// Pre-tag every line with the service name so logs are attributable
	// in a multi-service system.
	return l.WithField("service", service)
}
