package logger

import "go.uber.org/zap"

// MakeLogger инициализирует синглтон логера с необходимым уровнем логирования.
func MakeLogger(level string) (*zap.Logger, error) {
	var Log = zap.NewNop()

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	Log = zl
	return Log, nil
}
