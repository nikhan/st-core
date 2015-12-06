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
		graph: &Graph{},
	}
}

type Server struct {
	graph *Graph
}

func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) CreateElementsHandler(w http.ResponseWriter, r *http.Request) {
	element := context.Get(r, "body").([]CreateElement)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.Add(element, nil); err != nil {
		panic(err)
	}
}

func (s *Server) ParentCreateElementsHandler(w http.ResponseWriter, r *http.Request) {
	element := context.Get(r, "body").([]CreateElement)
	id := context.Get(r, "id").(ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.Add(element, &id); err != nil {
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

	if err := s.graph.Delete(id); err != nil {
		panic(err)
	}
}

func (s *Server) DeleteRecursiveElementHandler(w http.ResponseWriter, r *http.Request) {
	id := context.Get(r, "id").(ElementID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.DeleteRecursive(id); err != nil {
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

func (s *Server) RootExportGistHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Root/export")
}
func (s *Server) RootImportGistHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Root/import")
}
func (s *Server) ParentExportGistHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("parent/export")
}
func (s *Server) ParentImportGistHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("parent/import")
}
func (s *Server) LibraryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("library")
}

func (s *Server) TranslateElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "body").([]ID)

	s.graph.Lock()
	defer s.graph.Unlock()

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

	if err = s.graph.BatchTranslate(ids, x, y); err != nil {
		panic(err)
	}
}

func (s *Server) DeleteElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "body").([]ID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.BatchDelete(ids); err != nil {
		panic(err)
	}
}

func (s *Server) ResetElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "body").([]ID)

	s.graph.Lock()
	defer s.graph.Unlock()

	if err := s.graph.BatchReset(ids); err != nil {
		panic(err)
	}
}

func (s *Server) MoveElementsHandler(w http.ResponseWriter, r *http.Request) {
	ids := context.Get(r, "body").([]ID)

	s.graph.Lock()
	defer s.graph.Unlock()

	parent := ElementID(r.URL.Query().Get("parent"))
	// need error checking

	if err := s.graph.BatchMove(ids, parent); err != nil {
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
			"RootImportGist",
			"/pattern",
			"POST",
			[]string{"action", "gist"},
			s.RootImportGistHandler,
			[]Handler{RecoverHandler},
		},
		Endpoint{
			"RootExportGist",
			"/pattern",
			"GET",
			[]string{"action", "gist"},
			s.RootExportGistHandler,
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
			"ParentImportGist",
			"/pattern/{id}",
			"POST",
			[]string{"action", "gist"},
			s.ParentImportGistHandler,
			[]Handler{RecoverHandler},
		},
		Endpoint{
			"ParentExportGist",
			"/pattern/{id}",
			"GET",
			[]string{"action", "gist"},
			s.ParentExportGistHandler,
			[]Handler{RecoverHandler},
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
			"DeleteRecursiveElement",
			"/pattern/{id}",
			"DELETE",
			[]string{"action", "recursive"},
			s.DeleteRecursiveElementHandler,
			[]Handler{RecoverHandler, IdHandler},
		},
		Endpoint{
			"DeleteElement",
			"/pattern/{id}",
			"DELETE",
			[]string{},
			s.DeleteElementHandler,
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
			[]Handler{RecoverHandler, BatchHandler},
		},
		Endpoint{
			"DeleteElements",
			"/pattern",
			"PUT",
			[]string{"action", "delete"},
			s.DeleteElementsHandler,
			[]Handler{RecoverHandler, BatchHandler},
		},
		Endpoint{
			"MoveElements",
			"/pattern",
			"PUT",
			[]string{"action", "move"},
			s.MoveElementsHandler,
			[]Handler{RecoverHandler, BatchHandler},
		},
		Endpoint{
			"ResetElements",
			"/pattern",
			"PUT",
			[]string{"action", "reset"},
			s.ResetElementsHandler,
			[]Handler{RecoverHandler, BatchHandler},
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
