package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbpool, db_err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if db_err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to DB: %v\n", db_err)
		os.Exit(1)
	}
	defer dbpool.Close()
	router := chi.NewRouter()
	actionsRouter := chi.NewRouter()
	actionsRouter.Use(verifyApiKey)
	actionsRouter.Get("/info/{id}", redirectInfo(dbpool))
	actionsRouter.Post("/create", addRedirect(dbpool))
	actionsRouter.Put("/update/{id}", updateRedirect(dbpool))
	actionsRouter.Patch("/fix", patchRedirect(dbpool))
	actionsRouter.Delete("/disable/{id}", deleteRedirect(dbpool))
	operationsRouter := chi.NewRouter()
	operationsRouter.Use(verifyApiKey)
	operationsRouter.Get("/list", listall(dbpool))
	operationsRouter.Post("/searchPath", searchPath(dbpool))
	operationsRouter.Post("/destinationExists", redirectExists(dbpool))
	router.Use(middleware.Logger)
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
	log.Println("Sever running at Port 8082")
	log.Fatal(http.ListenAndServe("127.0.0.1:8082", router))
}
