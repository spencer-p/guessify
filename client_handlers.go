package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/zmb3/spotify"
)

type ClientHandler func(w http.ResponseWriter, r *http.Request, client spotify.Client)

type ClientHandlers struct {
	Clients map[string]spotify.Client
}

func (ch *ClientHandlers) NoClient(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, "%d: You are not logged in.", http.StatusForbidden)
}

func (ch *ClientHandlers) Whoami(w http.ResponseWriter, r *http.Request, client spotify.Client) {
	log.Print("calling whoami")

	// Get their name
	var name string
	user, err := client.CurrentUser()
	if err != nil {
		name = user.DisplayName
	} else {
		name = user.ID
	}

	fmt.Fprintf(w, "Hello %s!", name)
}

func (ch *ClientHandlers) Danceable(w http.ResponseWriter, r *http.Request, client spotify.Client) {
	log.Println("Getting danceable songs")
	fail := func(msg string, err error) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500: Internal error -- %s (%s)", msg, err)
	}

	// We will collect all the songs we can find in this slice
	var allSongs []spotify.ID

	// Start with library songs
	library, err := client.CurrentUsersTracks()
	if err != nil {
		fail("Could not read saved tracks", err)
		return
	}
	for _, track := range library.Tracks {
		allSongs = append(allSongs, track.ID)
	}
	log.Println("Got", len(library.Tracks), "tracks from the library")

	// Get the features and sort them by danceability
	features, err := client.GetAudioFeatures(allSongs...)
	if err != nil {
		fail("Could not get audio features", err)
		return
	}
	featureWrap := &SortableFeatures{features, library.Tracks}
	sort.Sort(featureWrap)
	log.Println("Sorted all the songs!")

	// Write the song list
	fmt.Fprintf(w, "<html><body>Here are your most danceable songs.\n<ol>")
	for _, track := range library.Tracks {
		fmt.Fprintf(w, "<li>%s</li>", track.Name)
	}
	fmt.Fprintf(w, "</ol></body></html>")
}

func (ch *ClientHandlers) WithClient(next ClientHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("userid")
		if err != nil {
			ch.NoClient(w, r)
			return
		}

		client, ok := ch.Clients[cookie.Value]
		if !ok {
			http.SetCookie(w, &http.Cookie{
				Name:   "userid",
				Value:  "INVALID",
				MaxAge: -1,
			})
			ch.NoClient(w, r)
			return
		}
		next(w, r, client)
	}
}

func (ch *ClientHandlers) HandleFuncs(mux *http.ServeMux) {
	mux.HandleFunc("/whoami", ch.WithClient(ch.Whoami))
	mux.HandleFunc("/danceable", ch.WithClient(ch.Danceable))
}

// Garbage to make the audio features sortable
type SortableFeatures struct {
	features []*spotify.AudioFeatures
	tracks   []spotify.SavedTrack
}

func (f *SortableFeatures) Len() int {
	return len(f.features)
}

func (f *SortableFeatures) Less(i, j int) bool {
	return f.features[i].Danceability < f.features[j].Danceability
}

func (f *SortableFeatures) Swap(i, j int) {
	f.features[i], f.features[j] = f.features[j], f.features[i]
	f.tracks[i], f.tracks[j] = f.tracks[j], f.tracks[i]
}
