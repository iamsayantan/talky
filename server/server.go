package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/iamsayantan/talky"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	chiware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/iamsayantan/talky/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebHandler interface {
	Route() chi.Router
	Authenticate(handler http.Handler) http.Handler
}

type Server struct {
	UserRepo store.UserRepository

	router chi.Router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(KeyAuthUser).(*talky.User)
	if !ok {
		errResp := struct {
			Error string `json:"error"`
		}{Error: "Invalid access token"}

		sendResponse(w, http.StatusBadRequest, errResp)
		return
	}

	log.Printf("Got Websocket Connection Request from User: %v", authUser)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("err: %v", err)
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}
		sendResponse(w, http.StatusNotAcceptable, errResp)
		return
	}

	ticker := time.NewTicker(6 * time.Second)
	for {
		select {
		case <-ticker.C:
			msg := struct {
				Ping string `json:"ping"`
			}{Ping: "ping"}

			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Error: %v", err)
				return
			}
		}
	}

}

func NewServer(userRepo store.UserRepository) *Server {
	s := &Server{
		UserRepo: userRepo,
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	})

	r := chi.NewRouter()
	r.Use(chiware.AllowContentType("application/json"))
	r.Use(corsHandler.Handler)

	h := NewUserHandler(s.UserRepo)
	r.Route("/user", func(r chi.Router) {
		r.Mount("/v1", h.Route())
	})

	r.Group(func(r chi.Router) {
		r.Use(h.Authenticate)
		r.Get("/ws", s.ServeWs)
	})

	s.router = r
	return s
}

func sendResponse(w http.ResponseWriter, statusCode int, v interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp, _ := json.Marshal(v)

	_, _ = w.Write(resp)
}
