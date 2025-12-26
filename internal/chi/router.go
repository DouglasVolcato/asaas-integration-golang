package chi

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const urlParamsKey contextKey = "chi_url_params"

type URLParams map[string]string

type route struct {
	method  string
	pattern string
	handler http.Handler
}

type Router interface {
	http.Handler
	Route(pattern string, fn func(r Router))
	Post(pattern string, handler http.HandlerFunc)
	Get(pattern string, handler http.HandlerFunc)
	Handle(pattern string, handler http.Handler)
	Use(middlewares ...func(http.Handler) http.Handler)
}

type Mux struct {
	prefix      string
	middlewares []func(http.Handler) http.Handler
	routes      *[]route
}

func NewRouter() *Mux {
	routes := []route{}
	return &Mux{routes: &routes}
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, rt := range *m.routes {
		if rt.method != "" && r.Method != rt.method {
			continue
		}
		if params, ok := matchPattern(rt.pattern, r.URL.Path); ok {
			ctx := context.WithValue(r.Context(), urlParamsKey, params)
			rt.handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}
	http.NotFound(w, r)
}

func (m *Mux) Route(pattern string, fn func(r Router)) {
	child := &Mux{
		prefix:      joinPath(m.prefix, pattern),
		middlewares: append([]func(http.Handler) http.Handler{}, m.middlewares...),
		routes:      m.routes,
	}
	fn(child)
}

func (m *Mux) Post(pattern string, handler http.HandlerFunc) {
	m.addRoute(http.MethodPost, pattern, handler)
}

func (m *Mux) Get(pattern string, handler http.HandlerFunc) {
	m.addRoute(http.MethodGet, pattern, handler)
}

func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.addRoute("", pattern, handler)
}

func (m *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *Mux) addRoute(method, pattern string, handler http.Handler) {
	path := joinPath(m.prefix, pattern)
	wrapped := handler
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		wrapped = m.middlewares[i](wrapped)
	}
	*m.routes = append(*m.routes, route{method: method, pattern: path, handler: wrapped})
}

func URLParam(r *http.Request, key string) string {
	params, _ := r.Context().Value(urlParamsKey).(URLParams)
	return params[key]
}

func matchPattern(pattern, path string) (URLParams, bool) {
	pattern = strings.TrimSuffix(pattern, "/")
	path = strings.TrimSuffix(path, "/")

	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(path, prefix) {
			return URLParams{}, true
		}
		return nil, false
	}

	pParts := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
	uParts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(pParts) != len(uParts) {
		return nil, false
	}

	params := URLParams{}
	for i := range pParts {
		if strings.HasPrefix(pParts[i], "{") && strings.HasSuffix(pParts[i], "}") {
			key := strings.TrimSuffix(strings.TrimPrefix(pParts[i], "{"), "}")
			params[key] = uParts[i]
			continue
		}
		if pParts[i] != uParts[i] {
			return nil, false
		}
	}
	return params, true
}

func joinPath(prefix, path string) string {
	if prefix == "" {
		return path
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}
