package stserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

func printGraph(g *Graph) {
	resp, _ := g.Get()
	o, _ := json.Marshal(resp)
	fmt.Println(string(o))
}

func makeRequest(router *mux.Router, method string, path string, body io.Reader) (*httptest.ResponseRecorder, error) {
	requestURL := fmt.Sprintf("http://localhost/%s", path)
	w := httptest.NewRecorder()
	r, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, err
	}
	router.ServeHTTP(w, r)
	return w, nil
}

func decode(w []byte, v interface{}) (interface{}, error) {
	t := reflect.TypeOf(v)
	val := reflect.New(t).Interface()
	err := json.NewDecoder(bytes.NewBuffer(w)).Decode(val)
	if err != nil {
		log.Println("received response:", string(w))
		return nil, err
	}
	return val, nil
}

func dump(filename string, left []byte) {
	ioutil.WriteFile(filename, left, 0644)
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

	// import all elements
	for _, s := range pattern {
		w, err := makeRequest(router, "POST", "pattern", bytes.NewBufferString(s))
		if err != nil || w.Code != 200 {
			t.Error("error with response", err)
		}
	}

	// set all values
	for _, s := range values {
		w, err := makeRequest(router, "PUT", fmt.Sprintf("pattern/%s", s.id), bytes.NewBufferString(s.body))
		if err != nil || w.Code != 200 {
			t.Error("error with response", err)
		}
	}

	// hide routes
	for _, s := range hide {
		w, err := makeRequest(router, "PUT", fmt.Sprintf("pattern/%s/route/%s", s.id, s.route), bytes.NewBufferString(s.body))
		if err != nil || w.Code != 200 {
			t.Error("error with response", err)
		}
	}

	// retrieve test_pattern element
	w, err := makeRequest(router, "GET", "pattern/test_pattern", nil)
	if err != nil || w.Code != 200 {
		t.Error("error with response", err)
	}

	body := w.Body.Bytes()

	// rerieve all elements
	w2, err := makeRequest(router, "GET", "pattern", nil)
	if err != nil || w.Code != 200 {
		t.Error("error with response", err)
	}

	body2 := w2.Body.Bytes()

	// compare, these should be the exactly the same
	if body == nil || !bytes.Equal(body, body2) {
		t.Error("exported bodies are not the same length")
	}

	list := []*CreateElement{}

	err = json.Unmarshal(body, &list)
	if err != nil {
		t.Error(err)
	}

	ids := url.Values{}
	ids.Set("action", "delete")
	for _, ce := range list {
		ids.Add("id", string(*ce.ID))
	}

	// delete all elements individually according to output order
	// TODO: posting NIL body to pattern/ causes panic!
	_, err = makeRequest(router, "PUT", "pattern?"+ids.Encode(), nil)
	if err != nil {
		t.Error(err)
	}

	zero, err := makeRequest(router, "GET", "pattern", nil)
	if err != nil {
		t.Error(err)
	}

	ce, _ := decode(zero.Body.Bytes(), []*CreateElement{})
	if len(*ce.(*[]*CreateElement)) > 0 {
		t.Error("delete did not remove all elements")
	}

	// post pattern back into streamtools
	_, err = makeRequest(router, "POST", "pattern", bytes.NewBuffer(body))
	if err != nil {
		t.Error(err)
	}

	// retrieve pattern
	w3, err := makeRequest(router, "GET", "pattern", nil)
	if err != nil {
		t.Error(err)
	}

	body3 := w3.Body.Bytes()

	// imported pattern should be same as original pattern
	if body3 == nil || !bytes.Equal(body, body3) {
		dump("left.json", body)
		dump("right.json", body3)
		t.Error("exported bodies are not the same length")
	}

	// get the 'init' group pattern
	init, err := makeRequest(router, "GET", "pattern/init", nil)
	if err != nil {
		t.Error(t)
	}

	//fmt.Println(string(init.Body.Bytes()))

	// delete the 'init' group from the server
	_, err = makeRequest(router, "PUT", "pattern?action=delete&id=init", nil)
	if err != nil {
		t.Error(err)
	}

	// TODO: ensure it has been deleted.
	_, err = makeRequest(router, "GET", "pattern", nil)
	if err != nil {
		t.Error(t)
	}

	// add the 'init' group back to the test_pattern group
	_, err = makeRequest(router, "POST", "pattern/test_pattern", init.Body)
	if err != nil {
		t.Error(t)
	}

	replacement := `{"id":"35","type":"connection","alias":"","source_id":"5","target_id":"7"}`
	_, err = makeRequest(router, "POST", "pattern", bytes.NewBufferString(replacement))
	if err != nil {
		t.Error(t)
	}

	w4, err := makeRequest(router, "GET", "pattern", nil)
	if err != nil {
		t.Error(t)
	}

	body4 := w4.Body.Bytes()

	// imported pattern should be same as original pattern
	if body4 == nil || !bytes.Equal(body, body4) {
		dump("body.json", body)
		dump("body4.json", body4)
		t.Error("exported bodies are not the same length")
	}

	//fmt.Println(string(p.Body.Bytes()))

	//ce, err := decode(p.Body.Bytes(), []*CreateElement{})
	//if err != nil {
	//	t.Error(t)
	//}

	//w.Body.Reset()
	//ce, err := decode(body, []*CreateElement{})
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println("??????", ce)

	/*w3, err := makeRequest(router, "POST", "pattern", w.Body)
	if err != nil {
		t.Error(err)
	}*/

	/*newPattern, err := decode(w, *[]*CreateElement{})
	if err != nil {
		t.Error(err)
	}*/

	//makeRequest(router, "GET", "pattern", nil)

	//makeRequest(router, "PUT", "pattern?action=delete&id=test_pattern", nil)
	//fmt.Println(reflect.TypeOf(w2.Body))
	//if w2 != w {
	//	t.Error("bodies not equal")
	//}

	//makeRequest(router, "POST", "pattern", w.Body)

	//makeRequest(router, "POST", "pattern?action=delete&id=test_pattern", nil)
}
