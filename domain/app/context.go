package app

import (
	"context"
	"log/slog"
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5/middleware"
)

type Context struct {
	context.Context
	IPAddress string
	RequestID uuid.UUID
	logger    *slog.Logger
}

func FromRequest(r *http.Request) Context {
	context := r.Context()
	if appContext, ok := context.(Context); ok {
		return appContext
	}

	appContext := Context{
		Context:   context,
		IPAddress: middleware.GetClientIP(context),
		RequestID: uuid.NewV4(),
	}

	return appContext
}

func (c Context) New() Context {
	return Context{
		Context:   context.Background(),
		IPAddress: c.IPAddress,
		RequestID: c.RequestID,
		logger:    c.logger,
	}
}

func (c Context) Logger() *slog.Logger {
	logger := c.logger

	if logger == nil {
		logger = slog.Default()
	}

	return logger.With("request_id", c.RequestID.String())
}
