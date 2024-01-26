package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sing3demons/test-kafka-log/logger"
)

type IMicroservice interface {
	StartHTTP()

	// HTTP Services
	GET(path string, h ServiceHandleFunc)
	POST(path string, h ServiceHandleFunc)
	PUT(path string, h ServiceHandleFunc)
	PATCH(path string, h ServiceHandleFunc)
	DELETE(path string, h ServiceHandleFunc)
}

type Microservice struct {
	*gin.Engine
	logger logger.ILogger
}

type ServiceHandleFunc func(c IContext)

func NewMicroservice(logger logger.ILogger) IMicroservice {

	r := gin.Default()
	// r.Use(logger.LoggingMiddleware())
	return &Microservice{r, logger}
}

func (ms *Microservice) GET(path string, handler ServiceHandleFunc) {
	ms.Engine.GET(path, func(ctx *gin.Context) {
		handler(NewContext(ms, ctx))
	})
}

func (ms *Microservice) POST(path string, handler ServiceHandleFunc) {
	ms.Engine.POST(path, func(ctx *gin.Context) {
		handler(NewContext(ms, ctx))
	})
}

func (ms *Microservice) PUT(path string, h ServiceHandleFunc) {
	ms.Engine.PUT(path, func(ctx *gin.Context) {
		h(NewContext(ms, ctx))
	})
}

func (ms *Microservice) PATCH(path string, h ServiceHandleFunc) {
	ms.Engine.PATCH(path, func(ctx *gin.Context) {
		h(NewContext(ms, ctx))
	})
}

func (ms *Microservice) DELETE(path string, handler ServiceHandleFunc) {
	ms.Engine.DELETE(path, func(ctx *gin.Context) {
		handler(NewContext(ms, ctx))
	})
}

func (ms *Microservice) StartHTTP() {
	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        ms,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		fmt.Printf("Listening and serving HTTP on %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}
