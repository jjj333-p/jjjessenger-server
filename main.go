package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	next_req_token      string
	messages_to_deliver []string
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
				next_req_token:      nt,
				messages_to_deliver: nil,
			}

			println(clients[string(body)].next_req_token)

		}

	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
