package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
			}
			fmt.Fprintf(w, "Received POST request with body: %s\n", string(body))
			println(string(body))
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
