package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func about(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("Hello, I Redirect URL's"))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(notFoundMessage))
}

func initDB() *pgxpool.Pool {
	dbpool, db_err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if db_err != nil {
		log.Fatalf("Unable to connect to DB: %v\n", db_err)
		defer os.Exit(1)
	}
	_, db_init_err1 := dbpool.Exec(context.Background(), urlredirectSchema)
	if db_init_err1 != nil {
		log.Fatalf("Error creating URL Redirects table: %v\n", db_init_err1)
		defer os.Exit(1)
	}
	_, db_init_err2 := dbpool.Exec(context.Background(), urlredirectAnalyticsSchema)
	if db_init_err2 != nil {
		log.Fatalf("Error creating URL Redirects Analytics table: %v\n", db_init_err2)
		defer os.Exit(1)
	}
	log.Println("DB initialized successfully")
	return dbpool
}

func initRouter(dbpool *pgxpool.Pool) *chi.Mux {
	router := chi.NewRouter()
	apiRouter := chi.NewRouter()
	router.Use(middleware.Heartbeat("/app/health"))
	router.Use(logRequest(dbpool))
	router.Use(httpRateLimit)
	router.Use(prometheusMiddleware)
	apiRouter.Use(verifyApiKey)
	router.Use(middleware.AllowContentType("application/json"))
	apiRouter.Get("/info/{id}", redirectInfo(dbpool))
	apiRouter.Post("/create", addRedirect(dbpool))
	apiRouter.Put("/update/{id}", updateRedirect(dbpool))
	apiRouter.Patch("/fix", patchRedirect(dbpool))
	apiRouter.Delete("/disable/{id}", deleteRedirect(dbpool))
	apiRouter.Get("/list", listall(dbpool))
	apiRouter.Post("/generate", generateRedirect(dbpool))
	apiRouter.Post("/search", searchPath(dbpool))
	apiRouter.Post("/check", redirectExists(dbpool))
	apiRouter.Post("/stats", stats(dbpool))
	router.Get("/*", handleRedirect(dbpool))
	router.Get("/notfound", notFound)
	router.Get("/about", about)
	router.NotFound(notFound)
	router.MethodNotAllowed(notFound)
	router.Mount("/redirector", apiRouter)
	router.Handle("/metrics", promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{}))
	return router
}

func main() {
	dbpool := initDB()
	defer dbpool.Close()
	initMetrics()
	router := initRouter(dbpool)
	log.Println("Sever running at Port 8082")
	log.Fatalln(http.ListenAndServe("127.0.0.1:8082", router))
}
