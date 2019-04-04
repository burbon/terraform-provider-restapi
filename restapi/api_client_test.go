package restapi

import (
	mylog "github.com/Mastercard/terraform-provider-restapi/log"
	"net/http"
	"testing"
	"time"
)

var api_client_server *http.Server

func TestAPIClient(t *testing.T) {
	log := mylog.New(debug)

	if debug {
		log.Println("client_test.go: Starting HTTP server")
	}
	setup_api_client_server()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:                   "http://127.0.0.1:8080/",
		insecure:              false,
		username:              "",
		password:              "",
		headers:               make(map[string]string, 0),
		timeout:               2,
		id_attribute:          "id",
		copy_keys:             make([]string, 0),
		write_returns_object:  false,
		create_returns_object: false,
		debug:                 debug,
	}
	client, err := NewAPIClient(opt)

	var res string

	if debug {
		log.Printf("api_client_test.go: Testing standard OK request\n")
	}
	res, err = client.send_request("GET", "/ok", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	if debug {
		log.Printf("api_client_test.go: Testing redirect request\n")
	}
	res, err = client.send_request("GET", "/redirect", "")
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify timeout works */
	if debug {
		log.Printf("api_client_test.go: Testing timeout aborts requests\n")
	}
	res, err = client.send_request("GET", "/slow", "")
	if err == nil {
		t.Fatalf("client_test.go: Timeout did not trigger on slow request")
	}

	if debug {
		log.Println("client_test.go: Stopping HTTP server")
	}
	shutdown_api_client_server()
	if debug {
		log.Println("client_test.go: Done")
	}
}

func setup_api_client_server() {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It works!"))
	})
	serverMux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(9999 * time.Second)
		w.Write([]byte("This will never return!!!!!"))
	})
	serverMux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusPermanentRedirect)
	})

	api_client_server = &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: serverMux,
	}
	go api_client_server.ListenAndServe()
	/* let the server start */
	time.Sleep(1 * time.Second)
}

func shutdown_api_client_server() {
	api_client_server.Close()
}
