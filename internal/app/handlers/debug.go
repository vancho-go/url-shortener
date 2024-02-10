package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/pprof"
)

func PprofHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", pprof.Index)
	r.Get("/cmdline", pprof.Cmdline)
	r.Get("/profile", pprof.Profile)
	r.Get("/symbol", pprof.Symbol)
	r.Get("/trace", pprof.Trace)
	r.Get("/allocs", pprof.Handler("allocs").ServeHTTP)
	r.Get("/block", pprof.Handler("block").ServeHTTP)
	r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.Get("/heap", pprof.Handler("heap").ServeHTTP)
	r.Get("/mutex", pprof.Handler("mutex").ServeHTTP)
	r.Get("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)

	return r
}
