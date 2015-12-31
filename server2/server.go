package stserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func NewServer() *Server {
	return &Server{
		graph: NewGraph(),
	}
}

type Server struct {
	graph *Graph
}

func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) CreateElementsHandler(w http.ResponseWriter, r *http.Request) {
	elements := context.Get(r, "body").(*[]*CreateElement)

	s.graph.Lock()
	defer s.graph.Unlock()

	ids, err := s.graph.Add(*elements, nil)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

func (s *Server) ParentCreateElementsHandler(w http.ResponseWriter, r *http.Request) {
	elements := context.Get(r, "body").(*[]*CreateElement)
	id := context.Get(r, "id").(ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if _, err := s.graph.Add(*elements, &id); err != nil {
		panic(err)
	}
}

func (s *Server) RootGetElementsHandler(w http.ResponseWriter, r *http.Request) {
	s.graph.Lock()
	defer s.graph.Unlock()

	element, err := s.graph.Get()

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(element)
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

func (s *Server) GetElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "body").([]ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	element, err := s.graph.Get(ids...)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(element)
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

	if err := s.graph.SetState(id, state); err != nil {
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

func (s *Server) UpdateGroupRouteHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)
	routeID := context.Get(r, "routeID").(ElementID)
	update := context.Get(r, "body").(*UpdateElement)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.UpdateGroupRoute(id, routeID, update); err != nil {
		panic(err)
	}
}

func (s *Server) LibraryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("library")
}

func (s *Server) TranslateElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "ids").([]ElementID)
	xs := context.Get(r, "x").(string)
	ys := context.Get(r, "y").(string)

	x, err := strconv.Atoi(xs)
	if err != nil {
		panic(err)
	}

	y, err := strconv.Atoi(ys)
	if err != nil {
		panic(err)
	}

	s.graph.Lock()
	defer s.graph.Unlock()

	if err = s.graph.BatchTranslate(ids, x, y); err != nil {
		panic(err)
	}
}

func (s *Server) DeleteElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "ids").([]ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.BatchDelete(ids); err != nil {
		panic(err)
	}
}

func (s *Server) ResetElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "ids").([]ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.BatchReset(ids); err != nil {
		panic(err)
	}
}

func (s *Server) UngroupElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "ids").([]ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.BatchUngroup(ids); err != nil {
		panic(err)
	}
}

func (s *Server) NewRouter() *mux.Router {
	type Handler func(http.Handler) http.Handler

	type Endpoint struct {
		Name        string
		Pattern     string
		Method      string
		Queries     []string
		HandlerFunc http.HandlerFunc
		Middle      []Handler
	}

	routes := []Endpoint{
		Endpoint{
			"WebSocket",
			"/ws",
			"GET",
			[]string{},
			s.WebSocketHandler,
			[]Handler{},
		},
		Endpoint{
			"library",
			"/library",
			"GET",
			[]string{},
			s.LibraryHandler,
			[]Handler{},
		},
		Endpoint{
			"RootGetElements",
			"/pattern",
			"GET",
			[]string{},
			s.RootGetElementsHandler,
			[]Handler{RecoverHandler},
		},
		Endpoint{
			"CreateElements",
			"/pattern",
			"POST",
			[]string{},
			s.CreateElementsHandler,
			[]Handler{RecoverHandler, CreateHandler},
		},
		Endpoint{
			"CreateElements",
			"/pattern/{id}",
			"POST",
			[]string{},
			s.ParentCreateElementsHandler,
			[]Handler{RecoverHandler, CreateHandler},
		},
		Endpoint{
			"GetElement",
			"/pattern/{id}",
			"GET",
			[]string{},
			s.GetElementHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"GetElementState",
			"/pattern/{id}/state",
			"GET",
			[]string{},
			s.GetElementStateHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"SetElementState",
			"/pattern/{id}/state",
			"PUT",
			[]string{},
			s.SetElementStateHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"UpdateElement",
			"/pattern/{id}",
			"PUT",
			[]string{},
			s.UpdateElementHandler,
			[]Handler{RecoverHandler, IdHandler, UpdateHandler},
		},
		Endpoint{
			"TranslateElements",
			"/pattern",
			"PUT",
			[]string{"action", "translate"},
			s.TranslateElementsHandler,
			[]Handler{RecoverHandler, QueryTranslateHandler},
		},
		Endpoint{
			"DeleteElements",
			"/pattern",
			"PUT",
			[]string{"action", "delete"},
			s.DeleteElementsHandler,
			[]Handler{RecoverHandler, QueryIDHandler},
		},
		Endpoint{
			"UngroupElements",
			"/pattern",
			"PUT",
			[]string{"action", "ungroup"},
			s.UngroupElementsHandler,
			[]Handler{RecoverHandler, QueryIDHandler},
		},
		Endpoint{
			"ResetElements",
			"/pattern",
			"PUT",
			[]string{"action", "reset"},
			s.ResetElementsHandler,
			[]Handler{RecoverHandler, QueryIDHandler},
		},
		Endpoint{
			"GetElements",
			"/pattern",
			"POST",
			[]string{"action", "get"},
			s.GetElementsHandler,
			[]Handler{RecoverHandler, QueryIDHandler},
		},
		Endpoint{
			"UpdateGroupRoute",
			"/pattern/{id}/route/{routeID}",
			"PUT",
			[]string{},
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
			Queries(route.Queries...).
			Handler(handler)
	}

	return router
}
