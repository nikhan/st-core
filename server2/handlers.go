package stserver

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

var (
	ErrBadRequest           = &Error{"bad_request", 400}
	ErrNotAcceptable        = &Error{"not_acceptable", 406}
	ErrUnsupportedMediaType = &Error{"unsupported_media_type", 415}
	ErrInternalServer       = &Error{"internal_server_error", 500}
)

type Errors struct {
	Errors []*Error `json:"errors"`
}

type Error struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

func VarHandler(inner http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		val, ok := vars[key]
		if !ok {
			WriteError(w, ErrNotAcceptable)
			return
		}

		if inner != nil {
			context.Set(r, key, val)
			inner.ServeHTTP(w, r)
		}
	})
}

func IdHandler(inner http.Handler) http.Handler {
	return VarHandler(inner, "id")
}

func RouteIdHandler(inner http.Handler) http.Handler {
	return VarHandler(inner, "routeID")
}

func BodyHandler(inner http.Handler, v interface{}) http.Handler {
	t := reflect.TypeOf(v)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := reflect.New(t).Interface()
		err := json.NewDecoder(r.Body).Decode(val)

		if err != nil {
			WriteError(w, ErrBadRequest)
			return
		}

		if inner != nil {
			context.Set(r, "body", val)
			inner.ServeHTTP(w, r)
		}
	})
}

func CreateHandler(inner http.Handler) http.Handler {
	return BodyHandler(inner, CreateElement{})
}

func UpdateHandler(inner http.Handler) http.Handler {
	return BodyHandler(inner, UpdateElement{})
}

func BatchHandler(inner http.Handler) http.Handler {
	return BodyHandler(inner, BatchElement{})
}

func RecoverHandler(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					log.Println("recoverhandler received non-error type")
					return
				}

				WriteError(w, &Error{err.Error(), 500})
			}
		}()

		inner.ServeHTTP(w, r)
	})
}

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
