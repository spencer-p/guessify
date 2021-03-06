package main

import (
	"fmt"
	"log"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<a href='/login'>Log in here</a>")
}

func login(w http.ResponseWriter, r *http.Request) {
	redirect := auth.AuthURL(STATE)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func redirectCallback(w http.ResponseWriter, r *http.Request) {
	// Check for token
	token, err := auth.Token(STATE, r)
	if err != nil {
		http.Error(w, "Failed to process Spotify token", http.StatusForbidden)
		log.Print("Could not process spotify token:", err)
	}

	// Check state is not stale
	if state := r.FormValue("state"); state != STATE {
		http.NotFound(w, r)
		log.Printf("State mismatched. Got %s, expected %s\n", state, STATE)
	}

	// Pull out a client with the token
	client := auth.NewClient(token)

	// Save the client with the userid
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal("Could not get user:", err)
	}
	clients[user.ID] = client

	// Put the client in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "userid",
		Value: user.ID,
	})

	http.Redirect(w, r, "/danceable", http.StatusFound)
}
