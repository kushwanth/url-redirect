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

func initRouter(dbpool *pgxpool.Pool) *chi.Mux {
	router := chi.NewRouter()
	actionsRouter := chi.NewRouter()
	operationsRouter := chi.NewRouter()
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
	router.Use(middleware.Logger)
	router.Use(httprate.Limit(10, time.Minute))
	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.Heartbeat("/app/health"))
	router.Get("/notfound", notFound)
	router.Get("/about", about)
	router.Get("/*", handleRedirect(dbpool))
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
	_, db_init_err := dbpool.Exec(context.Background(), urlredirectSchema)
	if db_init_err != nil {
		log.Fatalf("Error creating table: %v", db_init_err)
	}
	log.Println("DB initialized successfully")
	router := initRouter(dbpool)
	log.Println("Sever running at Port 8082")
	log.Fatal(http.ListenAndServe("127.0.0.1:8082", router))
}
