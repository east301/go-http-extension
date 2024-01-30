package httpext

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
)

type reverseProxyExSessionContextKeyType int

const ReverseProxyExSessionContextKey reverseProxyExSessionContextKeyType = iota

type ReverseProxyEx[T any] struct {
	OnRequest  func(*httputil.ProxyRequest) T
	OnComplete func(T, *http.Response, *http.Request) error
	OnError    func(T, http.ResponseWriter, *http.Request, error)
}

func (p *ReverseProxyEx[T]) handleRequest(request *httputil.ProxyRequest) {
	session := p.OnRequest(request)
	request.In = request.In.WithContext(context.WithValue(request.In.Context(), ReverseProxyExSessionContextKey, session))
	request.Out = request.Out.WithContext(context.WithValue(request.Out.Context(), ReverseProxyExSessionContextKey, session))
}

func (p *ReverseProxyEx[T]) handleResponse(response *http.Response) error {
	session, ok := response.Request.Context().Value(ReverseProxyExSessionContextKey).(T)
	if !ok {
		panic(errors.New("could not obtain session object"))
	}

	return p.OnComplete(session, response, response.Request)
}

func (p *ReverseProxyEx[T]) handleError(response http.ResponseWriter, request *http.Request, err error) {
	session, ok := request.Context().Value(ReverseProxyExSessionContextKey).(T)
	if !ok {
		panic(errors.New("could not obtain session object"))
	}

	p.OnError(session, response, request, err)
}

func (p *ReverseProxyEx[T]) AsHTTPHandler() http.Handler {
	return &httputil.ReverseProxy{
		Rewrite:        p.handleRequest,
		ModifyResponse: p.handleResponse,
		ErrorHandler:   p.handleError,
	}
}
