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

	"github.com/gorilla/websocket"
)

type idReq struct {
	id   string
	body string
}

type idRouteReq struct {
	id    string
	route string
	body  string
}

var smallPattern = []string{
	`[{"id":"test_first","type":"block","spec":"first"}]`,
	`[{"id":"test_latch","type":"block","spec":"latch"}]`,
	`[{"id":"test_group","type":"group","children":[{"id":"test_first"},{"id":"test_latch"}]}]`,
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

func makeRequest(host string, method string, path string, body io.Reader) (*http.Response, error) {
	requestURL := fmt.Sprintf("http://%s%s", host, path)
	r, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	return client.Do(r)
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

func findElement(id string, pattern interface{}) *Element {
	var element *Element
	ps := pattern.(*[]*Element)
	for _, ce := range *ps {
		if *ce.ID == ElementID(id) {
			element = ce
		}
	}
	return element
}

func dump(filename string, left []byte) {
	ioutil.WriteFile(filename, left, 0644)
}

func TestSmallPattern(t *testing.T) {
	s := NewServer()
	router := s.NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Error(err)
	}
	addr := serverURL.Host
	log.Printf("test server running: %s", addr)

	// import all elements
	for _, s := range smallPattern {
		w, err := makeRequest(addr, "POST", "/pattern", bytes.NewBufferString(s))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// retrieve test_pattern element
	w, err := makeRequest(addr, "GET", "/pattern/test_group", nil)
	if err != nil || w.StatusCode != 200 {
		t.Error("error with response", err)
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
	}

	expected := `[{"id":"1","type":"route","json_type":"any","direction":"input","name":"in"},{"id":"2","type":"route","json_type":"boolean","direction":"output","name":"first"},{"id":"test_first","type":"block","spec":"first","position":{"x":0,"y":0},"routes":[{"id":"1"},{"id":"2"}]},{"id":"3","type":"route","json_type":"any","direction":"input","name":"in"},{"id":"4","type":"route","json_type":"boolean","direction":"input","name":"ctrl"},{"id":"5","type":"route","json_type":"any","direction":"output","name":"true"},{"id":"6","type":"route","json_type":"any","direction":"output","name":"false"},{"id":"test_latch","type":"block","spec":"latch","position":{"x":0,"y":0},"routes":[{"id":"3"},{"id":"4"},{"id":"5"},{"id":"6"}]},{"id":"test_group","type":"group","position":{"x":0,"y":0},"routes":[{"id":"1","hidden":false,"alias":""},{"id":"2","hidden":false,"alias":""},{"id":"3","hidden":false,"alias":""},{"id":"4","hidden":false,"alias":""},{"id":"5","hidden":false,"alias":""},{"id":"6","hidden":false,"alias":""}],"children":[{"id":"test_first"},{"id":"test_latch"}]}]`

	// compare, these should be the exactly the same
	if body == nil || !bytes.Equal([]byte(expected), bytes.TrimSpace(body)) {
		t.Error("exported bodies are not the same length")
	}

}

func TestServer(t *testing.T) {
	s := NewServer()
	router := s.NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Error(err)
	}
	addr := serverURL.Host
	log.Printf("test server running: %s", addr)

	// import all elements
	for _, s := range pattern {
		w, err := makeRequest(addr, "POST", "/pattern", bytes.NewBufferString(s))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// set all values
	for _, s := range values {
		w, err := makeRequest(addr, "PUT", fmt.Sprintf("/pattern/%s", s.id), bytes.NewBufferString(s.body))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// retrieve test_pattern element
	w, err := makeRequest(addr, "GET", "/pattern/test_pattern", nil)
	if err != nil || w.StatusCode != 200 {
		t.Error("error with response", err)
	}

	tBody, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
	}

	test_pattern, err := decode(tBody, []*Element{})
	if err != nil {
		t.Error("could not decode body of test pattern, ", err)
	}

	elm := findElement("test_pattern", test_pattern)
	testPatternRoutesLength := len(elm.Routes)

	// hide routes
	for _, s := range hide {
		w, err := makeRequest(addr, "PUT", fmt.Sprintf("/pattern/%s/route/%s", s.id, s.route), bytes.NewBufferString(s.body))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// retrieve test_pattern element
	w, err = makeRequest(addr, "GET", "/pattern/test_pattern", nil)
	if err != nil || w.StatusCode != 200 {
		t.Error("error with response", err)
	}

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
	}

	test_pattern, err = decode(body, []*Element{})
	if err != nil {
		t.Error("could not decode body of test pattern, ", err)
	}

	// check to see if routes are hidden
	elm = findElement("test_pattern", test_pattern)
	if testPatternRoutesLength-10 != len(elm.Routes) {
		t.Error("hidden routes on test_pattern not hidden!")
		t.Error("should have ", testPatternRoutesLength-10, " has ", len(elm.Routes))
	}

	// check to see if alias is set
	inc := findElement("inc", test_pattern)
	found := false
	for _, r := range inc.Routes {
		if *r.ID == "9" && *r.Alias == "++" {
			found = true
		}
	}

	if !found {
		t.Error("alias could not be set on group!")
	}

	// check to see if our route had a value set
	route := findElement("21", test_pattern)
	if route.Value.Data != "250ms" {
		t.Error("value was not set!")
	}

	// rerieve all elements
	w2, err := makeRequest(addr, "GET", "/pattern", nil)
	if err != nil || w.StatusCode != 200 {
		t.Error("error with response", err)
	}

	body2, err := ioutil.ReadAll(w2.Body)
	if err != nil {
		t.Error(err)
	}

	// compare, these should be the exactly the same
	if body == nil || !bytes.Equal(body, body2) {
		t.Error("exported bodies are not the same length")
	}

	list := []*Element{}

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
	_, err = makeRequest(addr, "PUT", "/pattern?"+ids.Encode(), nil)
	if err != nil {
		t.Error(err)
	}

	zero, err := makeRequest(addr, "GET", "/pattern", nil)
	if err != nil {
		t.Error(err)
	}

	zbb, err := ioutil.ReadAll(zero.Body)
	if err != nil {
		t.Error(err)
	}

	ce, _ := decode(zbb, []*Element{})
	if len(*ce.(*[]*Element)) > 0 {
		fmt.Println(string(zbb))
		t.Error("delete did not remove all elements")
	}

	// post pattern back into streamtools
	_, err = makeRequest(addr, "POST", "/pattern", bytes.NewBuffer(body))
	if err != nil {
		t.Error(err)
	}

	// retrieve pattern
	w3, err := makeRequest(addr, "GET", "/pattern", nil)
	if err != nil {
		t.Error(err)
	}

	body3, err := ioutil.ReadAll(w3.Body)
	if err != nil {
		t.Error(err)
	}

	// TODO: make this work
	// imported pattern should be same as original pattern
	if body3 == nil || !bytes.Equal(body, body3) {
		dump("left.json", body)
		dump("right.json", body3)
		t.Error("exported bodies are not the same length")
	}

	// get the 'init' group pattern
	init, err := makeRequest(addr, "GET", "/pattern/init", nil)
	if err != nil {
		t.Error(t)
	}

	// delete the 'init' group from the server
	_, err = makeRequest(addr, "PUT", "/pattern?action=delete&id=init", nil)
	if err != nil {
		t.Error(err)
	}

	// TODO: ensure it has been deleted.
	_, err = makeRequest(addr, "GET", "/pattern", nil)
	if err != nil {
		t.Error(t)
	}

	// add the 'init' group back to the test_pattern group
	_, err = makeRequest(addr, "POST", "/pattern/test_pattern", init.Body)
	if err != nil {
		t.Error(t)
	}

	// add back the connection that was destroyed by deleting the init group
	replacement := `[{"id":"26","type":"connection","source_id":"5","target_id":"7"}]`
	_, err = makeRequest(addr, "POST", "/pattern", bytes.NewBufferString(replacement))
	if err != nil {
		t.Error(t)
	}

	// retrieve the pattern again
	w4, err := makeRequest(addr, "GET", "/pattern", nil)
	if err != nil {
		t.Error(t)
	}

	body4, err := ioutil.ReadAll(w4.Body)
	if err != nil {
		t.Error(err)
	}

	// imported pattern should be same as original pattern
	if body4 == nil || !bytes.Equal(body, body4) {
		dump("body.json", body)
		dump("body4.json", body4)
		t.Error("exported bodies are not the same length")
	}
}

func TestWebsocket(t *testing.T) {
	s := NewServer()
	router := s.NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Error(err)
	}
	addr := serverURL.Host
	log.Printf("test server running: %s", addr)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// import all elements
	for _, s := range pattern {
		w, err := makeRequest(addr, "POST", "/pattern", bytes.NewBufferString(s))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// set all values
	for _, s := range values {
		w, err := makeRequest(addr, "PUT", fmt.Sprintf("/pattern/%s", s.id), bytes.NewBufferString(s.body))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// hide routes
	for _, s := range hide {
		w, err := makeRequest(addr, "PUT", fmt.Sprintf("/pattern/%s/route/%s", s.id, s.route), bytes.NewBufferString(s.body))
		if err != nil || w.StatusCode != 200 {
			t.Error("error with response", err)
		}
	}

	// retrieve test_pattern element
	w, err := makeRequest(addr, "GET", "/pattern/test_pattern", nil)
	if err != nil || w.StatusCode != 200 {
		t.Error("error with response", err)
	}

	// subscribe to "test_group"
	c.WriteJSON(Element{ID: pElementID("test_pattern")})

	_, p, err := c.ReadMessage()

	expected := `{"action":"create","data":[{"id":"10","type":"route","json_type":"any","direction":"input","name":"in"},{"id":"27","type":"connection","source_id":"9","target_id":"10"},{"id":"11","type":"route","json_type":"string","direction":"input","name":"duration","value":{"data":"1s"}},{"id":"12","type":"route","json_type":"any","direction":"output","name":"out"},{"id":"28","type":"connection","source_id":"12","target_id":"7"},{"id":"7","type":"route","json_type":"number","direction":"input","name":"x"},{"id":"26","type":"connection","source_id":"5","target_id":"7"},{"id":"8","type":"route","json_type":"number","direction":"input","name":"y","value":{"data":1}},{"id":"9","type":"route","json_type":"number","direction":"output","name":"x+y"},{"id":"29","type":"connection","source_id":"9","target_id":"16"},{"id":"inc","type":"group","position":{"x":0,"y":0},"routes":[{"id":"10","hidden":false,"alias":""},{"id":"11","hidden":false,"alias":""},{"id":"12","hidden":false,"alias":""},{"id":"7","hidden":false,"alias":""},{"id":"8","hidden":true,"alias":""},{"id":"9","hidden":false,"alias":"++"}],"children":[{"id":"test_increment"},{"id":"test_increment_delay"}]},{"id":"1","type":"route","json_type":"any","direction":"input","name":"in","value":{"data":true}},{"id":"2","type":"route","json_type":"boolean","direction":"output","name":"first"},{"id":"25","type":"connection","source_id":"2","target_id":"4"},{"id":"3","type":"route","json_type":"any","direction":"input","name":"in"},{"id":"4","type":"route","json_type":"boolean","direction":"input","name":"ctrl"},{"id":"5","type":"route","json_type":"any","direction":"output","name":"true"},{"id":"6","type":"route","json_type":"any","direction":"output","name":"false"},{"id":"init","type":"group","position":{"x":0,"y":0},"routes":[{"id":"1","hidden":false,"alias":""},{"id":"2","hidden":false,"alias":""},{"id":"3","hidden":false,"alias":""},{"id":"4","hidden":false,"alias":""},{"id":"5","hidden":false,"alias":""},{"id":"6","hidden":false,"alias":""}],"children":[{"id":"test_first"},{"id":"test_latch"}]},{"id":"13","type":"route","json_type":"any","direction":"input","name":"trigger"},{"id":"33","type":"connection","source_id":"22","target_id":"13"},{"id":"14","type":"route","json_type":"any","direction":"output","name":"value"},{"id":"32","type":"connection","source_id":"14","target_id":"23"},{"id":"15","type":"route","json_type":"any","direction":"input","name":"\"value\"","source":"value"},{"id":"31","type":"link","source_id":"19","target_id":"15"},{"id":"16","type":"route","json_type":"any","direction":"input","name":"value"},{"id":"17","type":"route","json_type":"any","direction":"output","name":"out"},{"id":"34","type":"connection","source_id":"17","target_id":"24"},{"id":"18","type":"route","json_type":"any","direction":"input","name":"\"value\"","source":"value"},{"id":"30","type":"link","source_id":"19","target_id":"18"},{"id":"19","type":"route","json_type":"any","direction":"output","name":"value","source":"value"},{"id":"20","type":"route","json_type":"any","direction":"input","name":"in"},{"id":"21","type":"route","json_type":"string","direction":"input","name":"duration","value":{"data":"250ms"}},{"id":"22","type":"route","json_type":"any","direction":"output","name":"out"},{"id":"23","type":"route","json_type":"any","direction":"input","name":"log"},{"id":"24","type":"route","json_type":"any","direction":"input","name":"in"},{"id":"logger","type":"group","position":{"x":0,"y":0},"routes":[{"id":"13","hidden":true,"alias":""},{"id":"14","hidden":true,"alias":""},{"id":"15","hidden":true,"alias":""},{"id":"16","hidden":false,"alias":"save value"},{"id":"17","hidden":true,"alias":""},{"id":"18","hidden":true,"alias":""},{"id":"19","hidden":true,"alias":""},{"id":"20","hidden":false,"alias":""},{"id":"21","hidden":true,"alias":""},{"id":"22","hidden":true,"alias":""},{"id":"23","hidden":true,"alias":""},{"id":"24","hidden":false,"alias":""}],"children":[{"id":"test_delay_value"},{"id":"test_log"},{"id":"test_sink"},{"id":"test_value"},{"id":"test_value_get"},{"id":"test_value_set"}]},{"id":"test_pattern","type":"group","position":{"x":0,"y":0},"routes":[{"id":"1","hidden":false,"alias":""},{"id":"10","hidden":false,"alias":""},{"id":"11","hidden":false,"alias":""},{"id":"12","hidden":false,"alias":""},{"id":"16","hidden":false,"alias":""},{"id":"2","hidden":false,"alias":""},{"id":"20","hidden":false,"alias":""},{"id":"24","hidden":false,"alias":""},{"id":"3","hidden":false,"alias":""},{"id":"4","hidden":false,"alias":""},{"id":"5","hidden":false,"alias":""},{"id":"6","hidden":false,"alias":""},{"id":"7","hidden":false,"alias":""},{"id":"9","hidden":false,"alias":""}],"children":[{"id":"inc"},{"id":"init"},{"id":"logger"}]}]}`

	if !bytes.Equal([]byte(expected), p) {
		t.Error("invalid pattern returned by websocket")
	}

	// add a block to test pattern, this should fire a create event
	block := `[{"id":"+_pill","type":"block","spec":"+"}]`
	_, err = makeRequest(addr, "POST", "/pattern/test_pattern", bytes.NewBufferString(block))
	if err != nil {
		t.Error(t)
	}

	_, p, err = c.ReadMessage()

	expected = `{"action":"create","data":[{"id":"+_pill","type":"block","spec":"+","position":{"x":0,"y":0},"routes":[{"id":"35"},{"id":"36"},{"id":"37"}]},{"id":"35","type":"route","json_type":"number","direction":"input","name":"x"},{"id":"36","type":"route","json_type":"number","direction":"input","name":"y"},{"id":"37","type":"route","json_type":"number","direction":"output","name":"x+y"}]}`

	if !bytes.Equal([]byte(expected), p) {
		t.Error("invalid pattern returned by websocket")
	}

	ids := url.Values{}
	ids.Set("action", "delete")
	ids.Add("id", "+_pill")

	// delete all elements
	_, err = makeRequest(addr, "PUT", "/pattern?"+ids.Encode(), nil)
	if err != nil {
		t.Error(err)
	}

	_, p, err = c.ReadMessage()

	expected = `{"action":"delete","data":[{"id":"35"},{"id":"36"},{"id":"37"},{"id":"+_pill"}]}`

	if !bytes.Equal([]byte(expected), p) {
		t.Error("invalid pattern returned by websocket")
	}
}
