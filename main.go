package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	dataPrefix = []byte("data: ")
	dataSuffix = []byte("\n\n")
)

type bus struct {
	sync.RWMutex
	subs map[string]map[chan []byte]bool
}

func (b *bus) subscribe(channel string) (chan []byte, func()) {
	b.Lock()
	defer b.Unlock()

	sub := make(chan []byte)
	if _, ok := b.subs[channel]; !ok {
		b.subs[channel] = map[chan []byte]bool{}
	}
	b.subs[channel][sub] = true

	return sub, func() {
		b.Lock()
		defer b.Unlock()

		delete(b.subs[channel], sub)
		if len(b.subs[channel]) == 0 {
			delete(b.subs, channel)
		}
	}
}

func (b *bus) publish(channel string, data []byte) {
	b.RLock()
	defer b.RUnlock()

	for sub := range b.subs[channel] {
		sub <- data
	}
}

type server struct {
	b bus
}

func (s *server) get(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")

	sub, close := s.b.subscribe(req.URL.Path)
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

func (s *server) post(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	s.b.publish(req.URL.Path, data)
	w.WriteHeader(200)
}

func (s *server) handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch req.Method {
	case "GET":
		s.get(w, req)
	case "POST":
		s.post(w, req)
	default:
		w.WriteHeader(404)
	}
}

func main() {
	s := server{b: bus{subs: map[string]map[chan []byte]bool{}}}
	addr := ":" + os.Getenv("PORT")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(s.handler)))
}
