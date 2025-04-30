package logger

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, err error, fields ...Field)
}

type Field struct {
	Key   string
	Value interface{}
}

type ZerologService struct {
	logger zerolog.Logger
}

// Sensitive field names that should be masked
var sensitiveFields = []string{
	"password", "token", "secret", "key", "auth", "credential", "jwt", "body",
	"access_token", "refresh_token", "authorization", "api_key", "apikey",
}

// Function to sanitize sensitive data
func sanitizeValue(key string, value interface{}) interface{} {
	// If not a string value, return as is
	strValue, ok := value.(string)
	if !ok {
		return value
	}

	// Check if field name contains sensitive keywords
	keyLower := strings.ToLower(key)
	for _, field := range sensitiveFields {
		if strings.Contains(keyLower, field) {
			// Special handling for JSON bodies that might contain passwords
			if field == "body" && strings.Contains(strValue, "password") {
				// Mask passwords in JSON
				re := regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)
				strValue = re.ReplaceAllString(strValue, `"password":"*****"`)

				// Mask tokens in JSON
				re = regexp.MustCompile(`"token"\s*:\s*"[^"]*"`)
				strValue = re.ReplaceAllString(strValue, `"token":"*****"`)

				return strValue
			}

			// For other sensitive fields, mask completely
			return "*****"
		}
	}

	// Also check for JWT token patterns
	if strings.Contains(strValue, "eyJ") && strings.Count(strValue, ".") >= 2 {
		parts := strings.Split(strValue, ".")
		if len(parts) >= 3 {
			return parts[0][:10] + "..." // Only show the beginning of the token
		}
	}

	return value
}

func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: sanitizeValue(key, value)}
}

func NewLogger(level LogLevel) *ZerologService {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
		FormatLevel: func(i interface{}) string {
			var l string
			if ll, ok := i.(string); ok {
				switch ll {
				case "debug":
					l = "🔍  DEBUG"
				case "info":
					l = "🐳   INFO"
				case "warn":
					l = "🎃   WARN"
				case "error":
					l = "❌  ERROR"
				case "fatal":
					l = "💀  FATAL"
				default:
					l = "📝   LOG"
				}
			}
			return "\x1b[36m" + l + "\x1b[0m"
		},
		FormatMessage: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return "\x1b[32m" + i.(string) + "\x1b[0m"
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("\n    \x1b[34m%-20s:\x1b[0m", i)
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf(" \x1b[37m%v\x1b[0m", i)
		},
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
			zerolog.CallerFieldName,
		},
	}

	// Set global log level
	logLevel, err := zerolog.ParseLevel(string(level))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Create logger instance
	logger := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return &ZerologService{
		logger: logger,
	}
}

func (z *ZerologService) Debug(msg string, fields ...Field) {
	event := z.logger.Debug()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *ZerologService) Info(msg string, fields ...Field) {
	event := z.logger.Info()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *ZerologService) Warn(msg string, fields ...Field) {
	event := z.logger.Warn()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *ZerologService) Error(msg string, err error, fields ...Field) {
	event := z.logger.Error().Err(err)
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *ZerologService) Fatal(msg string, err error, fields ...Field) {
	event := z.logger.Fatal().Err(err)
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *ZerologService) addFields(event *zerolog.Event, fields ...Field) {
	for _, field := range fields {
		event.Interface(field.Key, field.Value)
	}
}
