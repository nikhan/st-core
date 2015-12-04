package stserver

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type ElementType int

type Elements interface {
	GetID()
	GetType()
}

type ID struct {
	ID string `json:"id"`
}

func (id *ID) GetID() string {
	return id.ID
}

type Element struct {
	ID
	Type  ElementType `json:"type"`
	Alias string      `json:"alias"`
}

func (e *Element) GetType() ElementType {
	return e.Type
}

type Core struct {
	Spec string `json:"spec"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Group struct {
	Element
	Position `json:"position"`
	Routes   []struct {
		ID
		Hidden bool `json:"hidden"`
		Label  bool `json:"label"`
	} `json:"routes"`
	Children []ID `json:"children"`
}

type Block struct {
	Element
	Core
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

type Source struct {
	Element
	Core
	Position `json:"position"`
	Routes   []ID `json:"routes"`
}

type Link struct {
	Element
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

type Connection struct {
	Element
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

type Route struct {
	Element
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Direction string      `json:"direction"`
	Source    string      `json:"source"`
}

type Server struct {
	sync.Mutex
	graph map[string]Elements
}

type Handler func(http.Handler) http.Handler

type Endpoint struct {
	Name        string
	Pattern     string
	Method      string
	HandlerFunc http.HandlerFunc
	Middle      []Handler
}

func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request)           {}
func (s *Server) CreateElementHandler(w http.ResponseWriter, r *http.Request)       {}
func (s *Server) GetElementHandler(w http.ResponseWriter, r *http.Request)          {}
func (s *Server) DeleteElementHandler(w http.ResponseWriter, r *http.Request)       {}
func (s *Server) GetElementStateHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) SetRouteValueHandler(w http.ResponseWriter, r *http.Request)       {}
func (s *Server) SetElementLabelHandler(w http.ResponseWriter, r *http.Request)     {}
func (s *Server) SetElementPositionHandler(w http.ResponseWriter, r *http.Request)  {}
func (s *Server) BatchModifyElementsHandler(w http.ResponseWriter, r *http.Request) {}
func (s *Server) SetGroupRouteHidden(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) SetGroupRouteAlias(w http.ResponseWriter, r *http.Request)         {}

func (s *Server) NewRouter() *mux.Router {
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
			[]Handler{},
		},
		Endpoint{
			"GetElement",
			"/pattern/{id}",
			"GET",
			s.GetElementHandler,
			[]Handler{},
		},
		Endpoint{
			"DeleteElement",
			"/pattern/{id}",
			"DELETE",
			s.DeleteElementHandler,
			[]Handler{},
		},
		Endpoint{
			"GetElementState",
			"/pattern/{id}/state",
			"GET",
			s.GetElementStateHandler,
			[]Handler{},
		},
		Endpoint{
			"SetElementValue",
			"/pattern/{id}/value",
			"PUT",
			s.SetRouteValueHandler,
			[]Handler{},
		},
		Endpoint{
			"SetElementLabel",
			"/pattern/{id}/label",
			"PUT",
			s.SetElementLabelHandler,
			[]Handler{},
		},
		Endpoint{
			"SetElementPosition",
			"/pattern/{id}/position",
			"PUT",
			s.SetElementPositionHandler,
			[]Handler{},
		},
		Endpoint{
			"BatchModifyElements",
			"/pattern/",
			"PUT",
			s.BatchModifyElementsHandler,
			[]Handler{},
		},
		Endpoint{
			"SetGroupRouteHidden",
			"/pattern/{id}/route/{routeID}/hidden",
			"PUT",
			s.SetGroupRouteHidden,
			[]Handler{},
		},
		Endpoint{
			"SetGroupRouteAlias",
			"/pattern/{id}/route/{routeID}/alias",
			"PUT",
			s.SetGroupRouteAlias,
			[]Handler{},
		},
	}
}

func (s *Server) Serve() error {

	return nil
}
