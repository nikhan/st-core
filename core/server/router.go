package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) NewRouter() *mux.Router {
	type Route struct {
		Name        string
		Pattern     string
		Method      string
		HandlerFunc http.HandlerFunc
	}

	routes := []Route{
		Route{
			"UpdateSocket",
			"/updates",
			"GET",
			s.UpdateSocket,
		},
		Route{
			"GroupIndex",
			"/groups",
			"GET",
			s.GroupIndex,
		},
		Route{
			"GroupCreate",
			"/groups",
			"POST",
			s.GroupCreate,
		},
		Route{
			"GroupDelete",
			"/groups/{id}",
			"DELETE",
			s.GroupDelete,
		},
		Route{
			"BlockIndex",
			"/blocks",
			"GET",
			s.BlockIndex,
		},
		Route{
			"BlockCreate",
			"/blocks",
			"POST",
			s.BlockCreate,
		},
		Route{
			"BlockDelete",
			"/blocks/{id}",
			"DELETE",
			s.BlockDelete,
		},
		Route{
			"BlockModifyName",
			"/blocks/{id}/name",
			"PUT",
			s.BlockModifyName,
		},
		Route{
			"BlockModifyRoute",
			"/blocks/{id}/routes/{index}",
			"PUT",
			s.BlockModifyRoute,
		},
		Route{
			"BlockModifyGroup",
			"/blocks/{id}/group",
			"PUT",
			s.BlockModifyGroup,
		},
		Route{
			"ConnectionIndex",
			"/connections",
			"GET",
			s.ConnectionIndex,
		},
		Route{
			"ConnectionCreate",
			"/connections",
			"POST",
			s.ConnectionCreate,
		},
		Route{
			"ConnectionDelete",
			"/connections/{id}",
			"DELETE",
			s.ConnectionDelete,
		},
		Route{
			"SourceCreate",
			"/sources",
			"POST",
			s.SourceCreate,
		},
		Route{
			"SourceIndex",
			"/sources",
			"GET",
			s.SourceIndex,
		},
		Route{
			"SourceModify",
			"/sources/{id}",
			"/PUT",
			s.SourceModify,
		},
		Route{
			"Source",
			"/sources/{id}",
			"GET",
			s.Source,
		},
		Route{
			"Source",
			"/sources/{id}",
			"DELETE",
			s.SourceDelete,
		},
	}
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	return router

}
