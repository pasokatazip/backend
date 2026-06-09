package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

//ミドルウェアつなげるための関数
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

//以下実際にrouterで呼び出す関数

func Auth(h http.Handler) http.Handler {
	return Chain(
		h,
		AuthMiddleware,
	)
}

func Premium(h http.Handler) http.Handler {
	return Chain(
		h,
		AuthMiddleware,
		RequireSubscMiddleware,
	)
}