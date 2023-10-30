package logger

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
// Никакой код навыка, кроме функции InitLogger, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var Log, _ = zap.NewDevelopment()

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	config := zap.NewDevelopmentConfig()
	parsedLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	config.Level = parsedLevel

	// устанавливаем синглтон
	Log, err = config.Build()
	if err != nil {
		return err
	}

	// Закрытие логгера
	defer Log.Sync()
	return nil
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processingStart := time.Now()
		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h(&lw, r)
		processingDuration := time.Since(processingStart)
		Log.Info("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("processing duration", processingDuration.String()),
			zap.String("status code", strconv.Itoa(responseData.status)),
			zap.String("response size", strconv.Itoa(responseData.size)),
		)
	}
}
