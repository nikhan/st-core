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
			context.Set(r, key, ElementID(val))
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
	return BodyHandler(inner, []*Element{})
}

func UpdateHandler(inner http.Handler) http.Handler {
	return BodyHandler(inner, Update{})
}

func QueryIDHandler(inner http.Handler) http.Handler {
	return BatchQueryHandler(inner)
}

func QueryTranslateHandler(inner http.Handler) http.Handler {
	return BatchQueryHandler(inner, "x", "y")
}

func BatchQueryHandler(inner http.Handler, keys ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, key := range keys {
			value, ok := r.URL.Query()[key]
			if !ok {
				WriteError(w, ErrBadRequest)
				return
			}

			if len(value) == 1 {
				context.Set(r, key, value[0])
			} else {
				context.Set(r, key, value)
			}
		}

		ids, ok := r.URL.Query()["id"]
		if !ok {
			WriteError(w, ErrBadRequest)
			return
		}

		eids := make([]ElementID, len(ids))
		for i, id := range ids {
			eids[i] = ElementID(id)
		}

		context.Set(r, "ids", eids)

		if inner != nil {
			inner.ServeHTTP(w, r)
		}
	})
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
				log.Println(identifyPanic())
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
			"%s\t%s\t%s",
			r.Method,
			name,
			time.Since(start),
		)
	})
}
