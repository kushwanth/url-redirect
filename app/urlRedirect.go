package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
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

func initRouter(dbpool *pgxpool.Pool) *chi.Mux {
	router := chi.NewRouter()
	actionsRouter := chi.NewRouter()
	operationsRouter := chi.NewRouter()
	router.Use(logRequest(dbpool))
	router.Use(httprate.Limit(10, time.Minute))
	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.Heartbeat("/app/health"))
	actionsRouter.Use(verifyApiKey)
	operationsRouter.Use(verifyApiKey)
	actionsRouter.Get("/info/{id}", redirectInfo(dbpool))
	actionsRouter.Post("/create", addRedirect(dbpool))
	actionsRouter.Put("/update/{id}", updateRedirect(dbpool))
	actionsRouter.Patch("/fix", patchRedirect(dbpool))
	actionsRouter.Delete("/disable/{id}", deleteRedirect(dbpool))
	operationsRouter.Get("/list", listall(dbpool))
	operationsRouter.Post("/generate", generateRedirect(dbpool))
	operationsRouter.Post("/searchPath", searchPath(dbpool))
	operationsRouter.Post("/destinationExists", redirectExists(dbpool))
	operationsRouter.Post("/stats", stats(dbpool))
	router.Get("/*", handleRedirect(dbpool))
	router.Get("/notfound", notFound)
	router.Get("/about", about)
	router.Mount("/api/action", actionsRouter)
	router.Mount("/api/operations", operationsRouter)
	router.NotFound(notFound)
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(errorMessage))
	})
	return router
}

func main() {
	dbpool, db_err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if db_err != nil {
		log.Fatalf("Unable to connect to DB: %v\n", db_err)
		os.Exit(1)
	}
	defer dbpool.Close()
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
	router := initRouter(dbpool)
	log.Println("Sever running at Port 8082")
	log.Fatalln(http.ListenAndServe("127.0.0.1:8082", router))
}
