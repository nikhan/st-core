package stserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func printGraph(g *Graph) {
	resp, _ := g.Get()
	o, _ := json.Marshal(resp)
	fmt.Println(string(o))
}

func makeRequest(router *mux.Router, method string, path string, body io.Reader) *httptest.ResponseRecorder {
	requestURL := fmt.Sprintf("http://localhost/%s", path)
	w := httptest.NewRecorder()
	fmt.Println(body, nil)
	r, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		log.Println(err)
		return w
	}
	router.ServeHTTP(w, r)
	fmt.Printf("---\ncode: %d\nbody: %s", w.Code, w.Body)
	return w
}

func TestServer(t *testing.T) {
	type idReq struct {
		id   string
		body string
	}

	type idRouteReq struct {
		id    string
		route string
		body  string
	}

	var pattern = []string{
		`[{"id":"test_first","type":"block","spec":"first"}]`,
		`[{"id":"test_latch","type":"block","spec":"latch"}]`,
		`[{"id":"test_increment","type":"block","spec":"+"}]`,
		`[{"id":"test_increment_delay","type":"block","spec":"delay"}]`,
		`[{"id":"test_value_get","type":"block","spec":"valueGet"}]`,
		`[{"id":"test_value_set","type":"block","spec":"valueSet"}]`,
		`[{"id":"test_value","type":"source","spec":"value"}]`,
		`[{"id":"test_delay_value","type":"block","spec":"delay"}]`,
		`[{"id":"test_log","type":"block","spec":"log"}]`,
		`[{"id":"test_sink","type":"block","spec":"sink"}]`,
		`[{"type":"connection","source_id":"2","target_id":"4"}]`,
		`[{"type":"connection","source_id":"5","target_id":"7"}]`,
		`[{"type":"connection","source_id":"9","target_id":"10"}]`,
		`[{"type":"connection","source_id":"12","target_id":"7"}]`,
		`[{"type":"connection","source_id":"9","target_id":"16"}]`,
		`[{"type":"link","source_id":"19","target_id":"18"}]`,
		`[{"type":"link","source_id":"19","target_id":"15"}]`,
		`[{"type":"connection","source_id":"14","target_id":"23"}]`,
		`[{"type":"connection","source_id":"22","target_id":"13"}]`,
		`[{"type":"connection","source_id":"17","target_id":"24"}]`,
		`[{"id":"init","type":"group","children":[{"id":"test_first"},{"id":"test_latch"}]}]`,
		`[{"id":"inc","type":"group","children":[{"id":"test_increment"},{"id":"test_increment_delay"}]}]`,
		`[{"id":"logger","type":"group","children":[{"id":"test_value_get"},{"id":"test_value_set"},{"id":"test_delay_value"},{"id":"test_value"},{"id":"test_log"},{"id":"test_sink"}]}]`,
		`[{"id":"test_pattern","type":"group","children":[{"id":"init"},{"id":"inc"},{"id":"logger"}]}]`,
	}

	var values = []idReq{
		idReq{
			id:   "1",
			body: `{"value":{"data":true}}`,
		},
		idReq{
			id:   "8",
			body: `{"value":{"data":1}}`,
		},
		idReq{
			id:   "11",
			body: `{"value":{"data":"1s"}}`,
		},
		idReq{
			id:   "21",
			body: `{"value":{"data":"250ms"}}`,
		},
	}

	var hide = []idRouteReq{
		idRouteReq{
			id:    "logger",
			route: "13",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "14",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "15",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "21",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "22",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "19",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "17",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "18",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "23",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "inc",
			route: "8",
			body:  `{"hidden":true}`,
		},
		idRouteReq{
			id:    "logger",
			route: "16",
			body:  `{"alias":"save value"}`,
		},
		idRouteReq{
			id:    "inc",
			route: "9",
			body:  `{"alias":"++"}`,
		},
	}

	server := NewServer()
	router := server.NewRouter()

	for _, s := range pattern {
		w := makeRequest(router, "POST", "pattern", bytes.NewBufferString(s))
		if w.Code != 200 {
			t.Error(w.Body)
		}
	}

	for _, s := range values {
		w := makeRequest(router, "PUT", fmt.Sprintf("pattern/%s", s.id), bytes.NewBufferString(s.body))
		if w.Code != 200 {
			t.Error(w.Body)
		}
	}

	for _, s := range hide {
		w := makeRequest(router, "PUT", fmt.Sprintf("pattern/%s/route/%s", s.id, s.route), bytes.NewBufferString(s.body))
		if w.Code != 200 {
			t.Error(w.Body)
		}
	}

	w := makeRequest(router, "GET", "pattern/test_pattern", nil)
	if w.Code != 200 {
		t.Error(w.Body)
	}

	//makeRequest(router, "POST", "pattern?action=delete&id=test_pattern", nil)
}
