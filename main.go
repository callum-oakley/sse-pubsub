package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	dataPrefix = []byte("data: ")
	dataSuffix = []byte("\n\n")
	subs       = map[string]map[chan []byte]bool{}
)

func subscribe(channel string) (chan []byte, func()) {
	sub := make(chan []byte)
	if _, ok := subs[channel]; !ok {
		subs[channel] = map[chan []byte]bool{}
	}
	subs[channel][sub] = true
	return sub, func() {
		delete(subs[channel], sub)
		if len(subs[channel]) == 0 {
			delete(subs, channel)
		}
	}
}

func publish(channel string, data []byte) {
	for sub := range subs[channel] {
		sub <- data
	}
}

func get(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")

	sub, close := subscribe(req.URL.Path)
	defer close()

	for data := range sub {
		event := []byte{}
		event = append(event, dataPrefix...)
		event = append(event, data...)
		event = append(event, dataSuffix...)
		if _, err := w.Write(event); err != nil {
			return
		}
		w.(http.Flusher).Flush()
	}
}

func post(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	publish(req.URL.Path, data)
	w.WriteHeader(200)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch req.Method {
	case "GET":
		get(w, req)
	case "POST":
		post(w, req)
	default:
		w.WriteHeader(404)
	}
}

func main() {
	addr := ":" + os.Getenv("PORT")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(handler)))
}
