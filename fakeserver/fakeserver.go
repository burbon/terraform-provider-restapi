package fakeserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Mastercard/terraform-provider-restapi/log"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type FakeServerOpts struct {
	Port    int
	Objects map[string]map[string]interface{}
	Start   bool
	Debug   bool
	Logger  *log.Logger
	Dir     string
}

type fakeserver struct {
	server  *http.Server
	objects map[string]map[string]interface{}
	debug   bool
	log     *log.Logger
	running bool
}

func NewFakeServer(opts *FakeServerOpts) *fakeserver {
	serverMux := http.NewServeMux()

	svr := &fakeserver{
		debug:   opts.Debug,
		log:     opts.Logger,
		objects: opts.Objects,
		running: false,
	}

	//If we were passed an argument for where to serve /static from...
	dir := opts.Dir
	if dir != "" {
		_, err := os.Stat(dir)
		if err == nil {
			svr.log.Printf("fakeserver.go: Will serve static files in '%s' under /static path", dir)
			serverMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
		} else {
			svr.log.Printf("fakeserver.go: WARNING: Not serving /static because directory '%s' does not exist", dir)
		}
	}

	finalHandler := http.HandlerFunc(svr.handle_api_object)
	serverMux.Handle("/", svr.middlewareDebug(finalHandler))

	api_object_server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", opts.Port),
		Handler: serverMux,
	}

	svr.server = api_object_server

	if opts.Start {
		svr.StartInBackground()
	}

	svr.log.Printf("fakeserver.go: Set up fakeserver: port=%d, debug=%t\n", opts.Port, opts.Debug)

	return svr
}

func (svr *fakeserver) StartInBackground() {
	go svr.server.ListenAndServe()

	/* Let the server start */
	time.Sleep(1 * time.Second)
	svr.running = true
}

func (svr *fakeserver) Shutdown() {
	svr.server.Close()
	svr.running = false
}

func (svr *fakeserver) Running() bool {
	return svr.running
}

func (svr *fakeserver) GetServer() *http.Server {
	return svr.server
}

func (svr *fakeserver) middlewareDebug(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/* Assume this will never fail */
		b, _ := ioutil.ReadAll(r.Body)
		ctx := context.WithValue(r.Context(), "b", b)

		if svr.debug {
			svr.log.Printf("fakeserver.go: Recieved request: %+v\n", r)
			svr.log.Printf("fakeserver.go: Headers:\n")
			for name, headers := range r.Header {
				name = strings.ToLower(name)
				for _, h := range headers {
					svr.log.Printf("fakeserver.go:  %v: %v", name, h)
				}
			}
			svr.log.Printf("fakeserver.go: BODY: %s\n", string(b))
			svr.log.Printf("fakeserver.go: IDs and objects:\n")
			for id, obj := range svr.objects {
				svr.log.Printf("  %s: %+v\n", id, obj)
			}
		}

		path := r.URL.EscapedPath()
		parts := strings.Split(path, "/")
		svr.log.Printf("fakeserver.go: Request received: %s %s\n", r.Method, path)
		svr.log.Printf("fakeserver.go: Split request up into %d parts: %v\n", len(parts), parts)
		if r.URL.RawQuery != "" {
			svr.log.Printf("fakeserver.go: Query string: %s\n", r.URL.RawQuery)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (svr *fakeserver) handle_api_object(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	if head != "api" {
		http.Error(w, fmt.Sprintf("Only /api is allowed, got: /%s", head), http.StatusBadRequest)
		return
	}
	head, r.URL.Path = ShiftPath(r.URL.Path)
	if head != "objects" {
		http.Error(w, fmt.Sprintf("Only /api/objects is allowed, got: /api/%s", head), http.StatusBadRequest)
		return
	}

	var id string
	id, r.URL.Path = ShiftPath(r.URL.Path)

	if r.URL.Path != "/" {
		http.Error(w, fmt.Sprintf("Unexpected extra parameters: %s, %s", head, r.URL.Path), http.StatusBadRequest)
		return
	}

	if id == "" {
		switch r.Method {
		case "GET":
			svr.handleGetList().ServeHTTP(w, r)
		case "POST":
			svr.handlePost().ServeHTTP(w, r)
		default:
			http.Error(w, "Only GET is allowed on collection", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case "GET":
		svr.handleGet(id).ServeHTTP(w, r)
	case "PUT":
		svr.handlePut(id).ServeHTTP(w, r)
	case "DELETE":
		svr.handleDelete(id).ServeHTTP(w, r)
	default:
		http.Error(w, "Only GET and PUT are allowed on object", http.StatusMethodNotAllowed)
	}
}

func (svr *fakeserver) handleGet(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		obj, ok := svr.objects[id]
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		svr.log.Printf("fakeserver.go: Returning object.\n")

		b, _ := json.Marshal(obj)
		w.Write(b)
	})
}
func (svr *fakeserver) handlePut(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svr.log.Printf("fakeserver.go: PUT")
		b := r.Context().Value("b").([]byte)

		svr.log.Printf("fakeserver.go: data sent - unmarshalling from JSON: %s\n", string(b))

		var obj map[string]interface{}
		if err := json.Unmarshal(b, &obj); err != nil {
			/* Failure goes back to the user as a 500. Log data here for
			   debugging (which shouldn't ever fail!) */
			svr.log.Printf("fakeserver.go: Unmarshal of request failed: %s\n", err)
			svr.log.Printf("\nBEGIN passed data:\n%s\nEND passed data.", string(b))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		/* Overwrite our stored test object */
		svr.log.Printf("fakeserver.go: Overwriting %s with new data:%+v\n", id, obj)
		svr.objects[id] = obj

		/* Coax the data we were sent back to JSON and send it to the user */
		b, _ = json.Marshal(obj)
		w.Write(b)
	})
}

func (svr *fakeserver) handleDelete(id string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svr.log.Printf("fakeserver.go: DELETE")
		delete(svr.objects, id)
		return
	})
}

func (svr *fakeserver) handleGetList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svr.log.Printf("fakeserver.go: GET list")
		result := make([]map[string]interface{}, 0)
		for _, hash := range svr.objects {
			result = append(result, hash)
		}
		b, _ := json.Marshal(result)
		w.Write(b)
	})
}

func (svr *fakeserver) handlePost() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svr.log.Printf("fakeserver.go: POST")
		b := r.Context().Value("b").([]byte)

		svr.log.Printf("fakeserver.go: data sent - unmarshalling from JSON: %s\n", string(b))

		var obj map[string]interface{}
		if err := json.Unmarshal(b, &obj); err != nil {
			/* Failure goes back to the user as a 500. Log data here for
			   debugging (which shouldn't ever fail!) */
			svr.log.Printf("fakeserver.go: Unmarshal of request failed: %s\n", err)
			svr.log.Printf("\nBEGIN passed data:\n%s\nEND passed data.", string(b))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// for i, idAttribute := range []string{"id", "Id", "ID"} {
		// }

		var id string
		if val, ok := obj["id"]; ok {
			id = fmt.Sprintf("%v", val)
		} else if val, ok := obj["Id"]; ok {
			id = fmt.Sprintf("%v", val)
		} else if val, ok := obj["ID"]; ok {
			id = fmt.Sprintf("%v", val)
		} else {
			svr.log.Printf("fakeserver.go: Bad request - POST to /api/objects without id field")
			http.Error(w, "POST sent with no id field in the data. Cannot persist this!", http.StatusBadRequest)
			return
		}

		_, ok := svr.objects[id]
		if ok {
			svr.log.Printf("fakeserver.go: Object exists. Allowing to overwrite: %s", id)
			// http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			// return
		}

		/* Creating new object */
		svr.log.Printf("fakeserver.go: Creating object %s with new data:%+v\n", id, obj)
		svr.objects[id] = obj

		/* Coax the data we were sent back to JSON and send it to the user */
		b, _ = json.Marshal(obj)
		w.Write(b)
	})
}

func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
