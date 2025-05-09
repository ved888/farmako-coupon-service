package server

import (
	"context"
	"farmako-coupon-service/middleware"
	"farmako-coupon-service/utils"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

type Server struct {
	chi.Router
	server *http.Server
}

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

// SetupBaseV1Routes provides all the routes that can be used
func SetupBaseV1Routes() *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestLoggerMiddleware)
	router.Use(middleware.CORSMiddleware())
	router.Route("/v1", func(v1 chi.Router) {
		// health endpoint
		v1.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondJSON(w, http.StatusOK, struct {
				Status string `json:"status"`
				Build  string `json:"build"`
			}{Status: "server is running", Build: utils.GetBuildNumber()})
		})

		// public
		v1.Route("/public", func(public chi.Router) {
			public.Group(PublicRoutes)
		})

		// admin routes
		v1.Route("/admin", func(admin chi.Router) {
			admin.Group(AdminRoutes)
		})

	})
	return &Server{Router: router}
}

func (svc *Server) Run(port string) error {
	svc.server = &http.Server{
		Addr:              port,
		Handler:           svc.Router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return svc.server.ListenAndServe()
}

func (svc *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return svc.server.Shutdown(ctx)
}
