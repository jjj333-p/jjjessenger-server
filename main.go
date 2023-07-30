package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	Next_req_token      *string
	Messages_to_deliver []string
}

func makeToken(length int) string {
	// Generate random bytes
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	// Encode as a string
	s := base64.StdEncoding.EncodeToString(b)

	return s
}

func syncToClient(body string, nt string, clients map[string]Client) {

	http.HandleFunc("/sync/"+nt, func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {

			// Set the Content-Type header to application/json
			w.Header().Set("Content-Type", "application/json")

			nt := makeToken((8))
			*clients[string(body)].Next_req_token = nt

			// Encode the object as JSON and write it to the response
			err := json.NewEncoder(w).Encode(clients[string(body)])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			go syncToClient(string(body), nt, clients)

		}

	})

}

func main() {

	clients := make(map[string]Client)

	//for testing
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {

			body, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
			}

			fmt.Fprintf(w, "Received POST request with body: %s\n", string(body))
			println(string(body))

		} else {

			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		}
	})

	http.HandleFunc("/getIdentity", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {

			body, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
			}

			nt := makeToken(8)

			clients[string(body)] = Client{
				Next_req_token:      &nt,
				Messages_to_deliver: nil,
			}

			// Encode the object as JSON and write it to the response
			er := json.NewEncoder(w).Encode(clients[string(body)])
			if er != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			go syncToClient(string(body), nt, clients)

		}

	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
