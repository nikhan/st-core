package stserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func NewServer() *Server {
	return &Server{
		graph: &Graph{},
	}
}

type Server struct {
	graph *Graph
}

func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) CreateElementHandler(w http.ResponseWriter, r *http.Request) {
	element := context.Get(r, "body").(*CreateElement)

	s.graph.Lock()
	defer s.graph.Unlock()

	err := s.graph.Add(element)

	if err != nil {
		panic(err)
	}
}

func (s *Server) GetElementHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	element, err := s.graph.Get(id)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(element)
}

func (s *Server) DeleteElementHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	err := s.graph.Delete(id)

	if err != nil {
		panic(err)
	}
}
func (s *Server) GetElementStateHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	state, err := s.graph.GetState(id)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}
func (s *Server) SetElementStateHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)
	state := struct{}{} // repalce with actual statehandler

	s.graph.Lock()
	defer s.graph.Unlock()

	err := s.graph.SetState(id, state)

	if err != nil {
		panic(err)
	}
}
func (s *Server) UpdateElementHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)
	update := context.Get(r, "body").(*UpdateElement)

	s.graph.Lock()
	defer s.graph.Unlock()

	err := s.graph.Update(id, update)

	if err != nil {
		panic(err)
	}
}
func (s *Server) UpdateElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "body").(*BatchElement)
	action := r.URL.Query().Get("action")

	s.graph.Lock()
	defer s.graph.Unlock()

	var err error
	switch action {
	case "translate":
		xs := r.URL.Query().Get("x")
		ys := r.URL.Query().Get("y")
		// need error checking

		x, err := strconv.Atoi(xs)
		if err != nil {
			panic(err)
		}

		y, err := strconv.Atoi(ys)
		if err != nil {
			panic(err)
		}

		err = s.graph.BatchTranslate(ids, x, y)
	case "move":
		parent := ElementID(r.URL.Query().Get("parent"))
		// need error checking

		err = s.graph.BatchMove(ids, parent)
	case "delete":
		err = s.graph.BatchDelete(ids)
	default:
		err = errors.New(fmt.Sprintf("unknown action: %s", action))
	}

	if err != nil {
		panic(err)
	}
}

func (s *Server) UpdateGroupRouteHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)
	routeID := context.Get(r, "routeID").(ElementID)

	update := context.Get(r, "body").(*UpdateElement)

	s.graph.Lock()
	defer s.graph.Unlock()

	err := s.graph.UpdateGroupRoute(id, routeID, update)

	if err != nil {
		panic(err)
	}
}

func (s *Server) ExportHandler(w http.ResponseWriter, r *http.Request) {}

func (s *Server) NewRouter() *mux.Router {
	type Handler func(http.Handler) http.Handler

	type Endpoint struct {
		Name        string
		Pattern     string
		Method      string
		HandlerFunc http.HandlerFunc
		Middle      []Handler
	}

	routes := []Endpoint{
		Endpoint{
			"WebSocket",
			"/ws",
			"GET",
			s.WebSocketHandler,
			[]Handler{},
		},
		Endpoint{
			"CreateElement",
			"/pattern",
			"POST",
			s.CreateElementHandler,
			[]Handler{RecoverHandler, CreateHandler},
		},
		Endpoint{
			"GetElement",
			"/pattern/{id}",
			"GET",
			s.GetElementHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"DeleteElement",
			"/pattern/{id}",
			"DELETE",
			s.DeleteElementHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"GetElementState",
			"/pattern/{id}/state",
			"GET",
			s.GetElementStateHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"GetElementState",
			"/pattern/{id}/state",
			"PUT",
			s.SetElementStateHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"UpdateElement",
			"/pattern/{id}",
			"PUT",
			s.UpdateElementHandler,
			[]Handler{RecoverHandler, IdHandler, UpdateHandler},
		},
		Endpoint{
			"UpdateElements",
			"/pattern/",
			"PUT",
			s.UpdateElementsHandler,
			[]Handler{RecoverHandler, BatchHandler},
		},
		Endpoint{
			"ExportElements",
			"/pattern/",
			"GET",
			s.ExportHandler,
			[]Handler{RecoverHandler},
		},
		Endpoint{
			"UpdateGroupRoute",
			"/pattern/{id}/route/{routeID}",
			"PUT",
			s.UpdateGroupRouteHandler,
			[]Handler{RecoverHandler, IdHandler, RouteIdHandler, UpdateHandler},
		},
	}

	router := mux.NewRouter()

	for _, route := range routes {
		var handle http.Handler
		handle = route.HandlerFunc

		for _, h := range route.Middle {
			handle = h(handle)
		}

		handler := Logger(handle, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
