// Package log is a log package.
package log

// Logger defines the interface, includes the supported logging methods.
// For each log level, it exposes four methods:
//
//   - methods named after the log level for log.Print-style logging
//   - methods ending in "w" for loosely-typed structured logging (read as "info with")
//   - methods ending in "f" for log.Printf-style logging
//   - methods ending in "ln" for log.Println-style logging
type Logger interface {
	Sync()

	Debug(args ...any)
	Debugf(template string, args ...any)
	Debugw(msg string, keysAndValues ...any)
	Debugln(args ...any)

	Info(args ...any)
	Infof(template string, args ...any)
	Infow(msg string, keysAndValues ...any)
	Infoln(args ...any)

	Warn(args ...any)
	Warnf(template string, args ...any)
	Warnw(msg string, keysAndValues ...any)
	Warnln(args ...any)

	Error(args ...any)
	Errorf(template string, args ...any)
	Errorw(msg string, keysAndValues ...any)
	Errorln(args ...any)

	Panic(args ...any)
	Panicf(template string, args ...any)
	Panicw(msg string, keysAndValues ...any)
	Panicln(args ...any)

	Fatal(args ...any)
	Fatalf(template string, args ...any)
	Fatalw(msg string, keysAndValues ...any)
	Fatalln(args ...any)
}
