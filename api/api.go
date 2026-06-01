package api

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Api struct {
	server *http.Server
}

func NewApi(mux *http.ServeMux) *Api {
	//Register routes and middlewares
	// wrap the entire mux with logger when serving
	log := Logger()

	mux.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	limiter := rate.NewLimiter(rate.Every(5*time.Second), 10)

	mux.Handle(
		"/limited",
		RateLimiter(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("limit ok"))
		}),
		),
	)

	// final handler will be the mux wrapped by logger
	return &Api{
		server: &http.Server{
			Handler: log(mux),
			Addr:    ":8000",
		},
	}

}

func (a *Api) Run() error {
	return a.server.ListenAndServe()
}
