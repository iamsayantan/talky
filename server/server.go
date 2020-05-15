package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	chiware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/iamsayantan/talky"
	"github.com/iamsayantan/talky/store"
	"log"
	"net/http"
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

	hub    *talky.Hub
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

	log.Printf("Got Websocket Connection Request from User: %d", authUser.ID)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("err: %v", err)
		errResp := struct {
			Error string `json:"error"`
		}{Error: err.Error()}
		sendResponse(w, http.StatusNotAcceptable, errResp)
		return
	}

	client := talky.NewClient(s.hub, authUser, conn)
	s.hub.AddClient(client)
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

	hub := talky.NewHub()

	s.router = r
	s.hub = hub
	return s
}

func sendResponse(w http.ResponseWriter, statusCode int, v interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp, _ := json.Marshal(v)

	_, _ = w.Write(resp)
}
