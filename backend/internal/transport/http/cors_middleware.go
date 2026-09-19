package http

import (
	"net/http"
	"strings"
)

// corsMiddleware answers cross origin requests for configured origins or wildcard.
// It sets permissive headers when configured with "*" or when the request origin matches.
func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigin, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")
			isAllowed := false

			for _, allowed := range origins {
				if allowed == "*" {
					isAllowed = true
					break
				}
				if origin != "" && strings.TrimRight(origin, "/") == strings.TrimRight(allowed, "/") {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				header := writer.Header()
				if allowedOrigin == "*" && origin == "" {
					header.Set("Access-Control-Allow-Origin", "*")
				} else if origin != "" {
					header.Set("Access-Control-Allow-Origin", origin)
				} else {
					header.Set("Access-Control-Allow-Origin", allowedOrigin)
				}
				header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Requested-With")
				header.Set("Access-Control-Max-Age", "86400")
				header.Add("Vary", "Origin")
			}

			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

// chain applies middleware in the order it is declared, so the first one in
// the list is the outermost.
func chain(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for index := len(middleware) - 1; index >= 0; index-- {
		handler = middleware[index](handler)
	}
	return handler
}
