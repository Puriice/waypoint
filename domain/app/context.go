package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type Context struct {
	context.Context
	IPAddress string
}

func FromRequest(r *http.Request) Context {
	context := r.Context()
	if appContext, ok := context.(Context); ok {
		return appContext
	}

	appContext := Context{
		Context:   context,
		IPAddress: middleware.GetClientIP(context),
	}

	return appContext
}
