package logger

import "go.uber.org/zap"

// New returns a zap logger configured for the given environment.
// "production" -> JSON output, Info level and above.
// anything else -> human-friendly console output, Debug level.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
