package stserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func toBuffer(i interface{}) bytes.Buffer {
	o, _ := json.Marshal(i)
	return bytes.NewBuffer(o)
}

func TestBlock(t *testing.T) {
	server := NewServer()
	router := server.NewRouter()
	/*`[{"type":"block"}]`
	`[{"type":"block","spec":"+"}]`
	`[{"type":"block","spec":"+","position":{"x":1,"y":1}}]`*/
	r, _ = http.NewRequest("GET", "/pattern", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	fmt.Printf("%+v\n", w.Code, w.Body)

}
