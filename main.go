package main

import (
	_ "fmt"
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/zmb3/spotify"
)

const (
	SHORTURL = "http://localhost"
	PORT     = ""
	URL      = SHORTURL + PORT
	REDIRECT = URL + "/callback"
	STATE    = "609a8184bd694ea3dd24bff8"
)

var (
	auth = spotify.NewAuthenticator(REDIRECT,
		spotify.ScopeUserLibraryRead,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistReadCollaborative,
		spotify.ScopeUserLibraryRead,
	)
	config  = struct{}{}
	clients = make(map[string]spotify.Client)
)

func main() {
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	log.Print("SPOTIFY_ID=" + os.Getenv("SPOTIFY_ID"))

	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/callback", redirectCallback)

	clientHandlers := &ClientHandlers{
		Clients: clients,
	}
	clientHandlers.HandleFuncs(mux)

	server := &http.Server{
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
