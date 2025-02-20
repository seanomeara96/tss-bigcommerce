package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Custom handler type definition
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handler wrapper
type Handler struct {
	h      HandlerFunc
	logger *zap.Logger
}

// Context key type for request ID
type contextKey string

const requestIDKey contextKey = "requestID"

// ServeHTTP implements the http.Handler interface
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	ctx := context.WithValue(r.Context(), requestIDKey, requestID)
	r = r.WithContext(ctx)

	if err := h.h(w, r); err != nil {
		h.logger.Error("Request failed",
			zap.Error(err),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		go sendTelegramNotification(fmt.Sprintf("Server Error: %v (Request ID: %s)", err, requestID))
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}
}

// Send notification via Telegram
func sendTelegramNotification(message string) error {
	// Replace with your bot's API token
	YOUR_BOT_API_TOKEN := os.Getenv("TELEGRAM_BOT_API_TOKEN")
	if YOUR_BOT_API_TOKEN == "" {
		return errors.New("TELEGRAM_BOT_API_TOKEN env var is empty")
	}
	botToken := YOUR_BOT_API_TOKEN
	// Replace with the chat ID you want to send a message to
	YOUR_CHAT_ID := os.Getenv("TELEGRAM_CHAT_ID")
	if YOUR_CHAT_ID == "" {
		return errors.New("TELEGRAM_CHAT_ID env var is empty")
	}

	_chatID, err := strconv.Atoi(YOUR_CHAT_ID)
	if err != nil {
		return fmt.Errorf("could not convert TELEGRAM_CHAT_ID var to int. %w", err)
	}

	chatID := int64(_chatID)

	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("unable to create bot instance. %w", err)
	}

	// Create a new message
	msg := tgbotapi.NewMessage(chatID, message)

	// Send the message
	_, err = bot.Send(msg)
	if err != nil {
		zap.L().Error("Failed to send Telegram notification", zap.Error(err))
		return errors.New("failed to send Telegram notification")
	}

	return nil
}

// Logging middleware with zap
func loggingMiddleware(logger *zap.Logger, next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		start := time.Now()
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		requestID := r.Context().Value(requestIDKey).(string)

		err := next(lw, r)
		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", lw.statusCode),
			zap.Duration("duration", duration),
		}

		if err != nil {
			logger.Error("Request completed with error", append(fields, zap.Error(err))...)
		} else {
			logger.Info("Request completed", fields...)
		}

		return err
	}
}

// Custom ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

// Example handler function
func helloHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}

	_, err := fmt.Fprintf(w, "Hello, World!\n")
	if err != nil {
		return fmt.Errorf("failed to write response: %v", err)
	}

	return nil
}

// Setup production-ready logger
func setupLogger() *zap.Logger {
	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Configure log file rotation
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:      "json",
		EncoderConfig: encoderConfig,
		OutputPaths: []string{
			"stdout",
			"server.log",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Make logger globally available
	zap.ReplaceGlobals(logger)
	return logger
}

func main() {
	// Initialize logger
	logger := setupLogger()
	defer logger.Sync() // Flush logs on exit

	// Create handler with logging middleware
	handler := Handler{
		h:      loggingMiddleware(logger, helloHandler),
		logger: logger,
	}

	http.Handle("/", handler)

	port := ":8080"
	logger.Info("Server starting", zap.String("port", port))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logger.Error("Server failed to start", zap.Error(err))
		os.Exit(1) // Explicitly exit after logging, allowing defer to run
	}
}
