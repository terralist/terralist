package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_logger "gorm.io/gorm/logger"
)

type Logger struct{}

func (l *Logger) LogMode(_logger.LogLevel) _logger.Interface {
	return l
}

func (l *Logger) Error(ctx context.Context, msg string, opts ...interface{}) {
	log.Error().Msg(fmt.Sprintf(msg, opts...))
}

func (l *Logger) Warn(ctx context.Context, msg string, opts ...interface{}) {
	log.Warn().Msg(fmt.Sprintf(msg, opts...))
}

func (l *Logger) Info(ctx context.Context, msg string, opts ...interface{}) {
	log.Info().Msg(fmt.Sprintf(msg, opts...))
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, f func() (string, int64), err error) {
	var event *zerolog.Event

	if err != nil {
		event = log.Debug()
	} else {
		event = log.Trace()
	}

	var dur_key string

	switch zerolog.DurationFieldUnit {
	case time.Nanosecond:
		dur_key = "elapsed_ns"
	case time.Microsecond:
		dur_key = "elapsed_us"
	case time.Millisecond:
		dur_key = "elapsed_ms"
	case time.Second:
		dur_key = "elapsed"
	case time.Minute:
		dur_key = "elapsed_min"
	case time.Hour:
		dur_key = "elapsed_hr"
	default:
		log.Error().
			Interface("zerolog.DurationFieldUnit", zerolog.DurationFieldUnit).
			Msg("Unknown value for DurationFieldUnit")
		dur_key = "elapsed_"
	}

	event.Dur(dur_key, time.Since(begin))

	sql, rows := f()
	if sql != "" {
		event.Str("sql", sql)
	}
	if rows > -1 {
		event.Int64("rows", rows)
	}

	event.Send()
}
