package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Client struct {
	NextReqToken      string
	MessagesToDeliver []string
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

var clients = make(map[string]*Client)

var syncTokens = make(map[string]string)

var tmpTokenLength = 8

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

func SyncToClient(hw http.ResponseWriter, hr *http.Request) {

	//pull the token from the url
	syncToken := chi.URLParam(hr, "syncToken")

	//fetch the identity from the map
	identity := syncTokens[syncToken]

	//purge that token's entry from the map for memory efficency
	delete(syncTokens, syncToken)

	nt := makeToken(tmpTokenLength)

	//if couldnt find an identity for that token its unauthed
	if clients[identity] == nil {

		hw.Write([]byte("401 Unauthorized"))

		return

	}

	clients[identity].NextReqToken = nt

	//make temporary sync token point to the object with the identity
	syncTokens[nt] = identity

	// Create a buffer to hold the encoded data
	buf := new(bytes.Buffer)

	// Create a new encoder and encode the client obj
	enc := gob.NewEncoder(buf)
	err := enc.Encode(clients[identity])
	if err != nil {
		panic(err)
	}

	//return the object
	hw.Write(buf.Bytes())

	//clear the queue of messages
	clients[identity].MessagesToDeliver = []string{}

}

func getIdentity(hw http.ResponseWriter, hr *http.Request) {

	//internal server identity
	identity := chi.URLParam(hr, "pubKey")

	//sync token
	nt := makeToken(tmpTokenLength)

	//create new client with that identity
	client := Client{
		NextReqToken:      nt,
		MessagesToDeliver: nil,
	}
	clients[identity] = &client

	//make temporary sync token point to the object with the identity
	syncTokens[nt] = identity

	//respond to get request with the next sync token
	hw.Write([]byte(nt))

}

func main() {

	r := chi.NewRouter()

	//log for testing
	r.Use(middleware.Logger)

	//help handle errors?
	r.Use(middleware.Recoverer)

	//idk
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	//dont need to nested call, will just store based on keys in map
	r.Get("/sync/{syncToken}", SyncToClient)

	//create new identity with the server
	r.Get("/identity/{pubKey}", getIdentity)

	http.ListenAndServe(":8080", r)

}
