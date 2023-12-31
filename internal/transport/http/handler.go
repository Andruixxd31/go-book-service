package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)


type Handler struct {
    Router *mux.Router
    BooksService BooksService
    AccountsService AccountsService
    Server *http.Server
}


type Response struct {
    Status int
    Message string
    Count int
}

func NewHandler(booksService BooksService, accountsService AccountsService) *Handler {
    h := &Handler{
        BooksService: booksService,
        AccountsService: accountsService,
    }
    h.Router = mux.NewRouter()

    h.mapRoutes()
    h.Router.Use(JSONMiddleware)
    h.Router.Use(LoggingMiddleWare)
    h.Router.Use(TimeOutMiddleware)

    h.Server = &http.Server{
        Addr: "0.0.0.0:8080",
        Handler: h.Router,
    }
    return h
}


func (h *Handler) mapRoutes(){
    h.Router.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "I am alive")
    })

    h.Router.HandleFunc("/api/v1/account", JWTAuth(h.CreateAccount)).Methods("POST")
    h.Router.HandleFunc("/api/v1/account/{id}", h.GetAccount).Methods("GET")
    h.Router.HandleFunc("/api/v1/account/{id}", JWTAuth(h.UpdateAccount)).Methods("PATCH")
    h.Router.HandleFunc("/api/v1/account/{id}", JWTAuth(h.DeleteAccount)).Methods("DELETE")

    h.Router.HandleFunc("/api/v1/book", JWTAuth(h.CreateBook)).Methods("POST")
    h.Router.HandleFunc("/api/v1/book/{id}", h.GetBook).Methods("GET")
    h.Router.HandleFunc("/api/v1/book/{id}", JWTAuth(h.UpdateBook)).Methods("PATCH")
    h.Router.HandleFunc("/api/v1/book/{id}", JWTAuth(h.DeleteBook)).Methods("DELETE")
    h.Router.HandleFunc("/api/v1/book/upvote-count/{id}", JWTAuth(h.GetUpVoteCount)).Methods("GET")
    h.Router.HandleFunc("/api/v1/book/upvote/{accountid}/{id}", h.UpVoteBook).Methods("PATCH")
    h.Router.HandleFunc("/api/v1/book/downvote/{accountid}/{id}", JWTAuth(h.DownVoteBook)).Methods("PATCH")
}

func (h *Handler) Serve() error {
    go func() {
        if err := h.Server.ListenAndServe(); err != nil {
            log.Println(err.Error())
        }
    }()

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <- c

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    log.Println("shut down gracefully")
    h.Server.Shutdown(ctx)

    return nil
}
