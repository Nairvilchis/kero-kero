package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger es el logger global de la aplicación
var Logger zerolog.Logger

// Init inicializa el sistema de logging
func Init(level, format string) {
	// Configurar nivel de log
	logLevel := zerolog.InfoLevel
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Configurar formato
	if format == "text" {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02T15:04:05Z07:00"}
		if logLevel == zerolog.DebugLevel {
			Logger = log.Output(output).With().Caller().Logger()
		} else {
			Logger = log.Output(output)
		}
	} else {
		if logLevel == zerolog.DebugLevel {
			Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
		} else {
			Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		}
	}

	log.Logger = Logger
}

// Debug registra un mensaje de debug
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Info registra un mensaje informativo
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Warn registra una advertencia
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Error registra un error
func Error(msg string, err error) {
	Logger.Error().Err(err).Msg(msg)
}

// Fatal registra un error fatal y termina la aplicación
func Fatal(msg string, err error) {
	Logger.Fatal().Err(err).Msg(msg)
}

// With retorna un logger con campos adicionales
func With() zerolog.Context {
	return Logger.With()
}
