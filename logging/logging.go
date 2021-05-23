package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger = func() *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		os.Stderr,
		zap.DebugLevel,
	)
	return zap.New(core)
}()

// New returns a logger for package.
func New(pkg string) *zap.Logger {
	return logger.Named(pkg)
}
