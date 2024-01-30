package httpext

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
)

type reverseProxyExSessionContextKeyType int

const reverseProxyExSessionContextKey reverseProxyExSessionContextKeyType = iota

type ReverseProxyExHandler[T any] interface {
	OnRequest(*httputil.ProxyRequest) T
	OnComplete(T, *http.Response, *http.Request) error
	OnError(T, http.ResponseWriter, *http.Request, error)
}

type ReverseProxyEx[T any] struct {
	handler ReverseProxyExHandler[T]
}

func (p *ReverseProxyEx[T]) handleRequest(request *httputil.ProxyRequest) {
	session := p.handler.OnRequest(request)
	request.In = request.In.WithContext(context.WithValue(request.In.Context(), reverseProxyExSessionContextKey, session))
	request.Out = request.Out.WithContext(context.WithValue(request.Out.Context(), reverseProxyExSessionContextKey, session))
}

func (p *ReverseProxyEx[T]) handleResponse(response *http.Response) error {
	session, ok := response.Request.Context().Value(reverseProxyExSessionContextKey).(T)
	if !ok {
		return errors.New("could not obtain session object")
	}

	return p.handler.OnComplete(session, response, response.Request)
}

func (p *ReverseProxyEx[T]) handleError(response http.ResponseWriter, request *http.Request, err error) {
	session, ok := request.Context().Value(reverseProxyExSessionContextKey).(T)
	if !ok {
		return
	}

	p.handler.OnError(session, response, request, err)
}

func (p *ReverseProxyEx[T]) AsHTTPHandler() http.Handler {
	return &httputil.ReverseProxy{
		Rewrite:        p.handleRequest,
		ModifyResponse: p.handleResponse,
		ErrorHandler:   p.handleError,
	}
}

func NewReverseProxyEx[T any](handler ReverseProxyExHandler[T]) *ReverseProxyEx[T] {
	return &ReverseProxyEx[T]{handler: handler}
}
