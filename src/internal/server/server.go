package server

import (
	"context"
	"database/sql"
	"errors"
	"feedscollector/internal"
	"feedscollector/internal/infrastructure/config"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

// API struct holds the database connection and router
type API struct {
	DB *sql.DB
}

func NewAPI(db *sql.DB) *API {
	return &API{DB: db}
}

func RunAPIServer(ctxWithCancel context.Context, db *sql.DB, config *config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Println("Starting web server on :", config.Server.Port)
	apiInstance := NewAPI(db)

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiInstance.RegisterRoutes(apiRouter)
	http.Handle("/", AddCORSHeaders(router))

	server := &http.Server{
		Addr:              ":" + config.Server.Port,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		<-ctxWithCancel.Done()
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 4*time.Second)
		defer shutdownRelease()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			internal.ErrorLogger.Fatalf("Error shutting down server: %v", err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		internal.ErrorLogger.Fatalf("Error starting server: %v", err)
	}
}

func (api *API) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/channels", api.ListChannels).Methods("GET")
	router.HandleFunc("/channels", api.AddChannel).Methods("POST")
	router.HandleFunc("/channels/{id}", api.UpdateChannel).Methods("PUT")
	router.HandleFunc("/channels/{id}", api.PatchChannel).Methods("PATCH")
	router.HandleFunc("/channels/{id}", api.DeleteChannel).Methods("DELETE")
	router.HandleFunc("/channels/{id}/items", api.listItems).Methods("GET")
	router.HandleFunc("/channels/{channel_id}/items/{item_id}", api.RemoveItemFromChannel).Methods("DELETE")
	router.HandleFunc("/items/{id}", api.PatchItem).Methods("PATCH")
	router.HandleFunc("/items/{id}", api.DeleteItem).Methods("DELETE")
	// TODO: router.HandleFunc("/tags", api.ListTags).Methods("GET")
	// TODO: add tag to channel
	// TODO: remove tag from channel
	// TODO: add tag to item
	// TODO: remove tag from item
	router.HandleFunc("/groups", api.ListGroups).Methods("GET")
	router.HandleFunc("/groups", api.AddGroup).Methods("POST")
	router.HandleFunc("/groups/{id}", api.UpdateGroup).Methods("PUT")
	router.HandleFunc("/groups/{id}", api.DeleteGroup).Methods("DELETE")
	router.HandleFunc("/groups", api.AddChannelToGroup).Methods("POST")
	router.HandleFunc("/groups", api.RemoveChannelFromGroup).Methods("DELETE")
}
