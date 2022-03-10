package main

import (
	"context"
	"log"
	"net/http"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/suryatresna/ws-sample/internal/pkg/epoller"
	"github.com/suryatresna/ws-sample/internal/pkg/socket"
)

func main() {
	log.Println("Starting services")
	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	epoll, err := epoller.NewPollerWithBuffer(1000)
	if err != nil {
		log.Println("[ERR] fail ini newPoller")
		return
	}

	log.Println("initialize NewSocket")
	sock := socket.NewSocket(epoll)

	go sock.Listen()
	go sock.Start()

	r.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Query().Get("room")
		if roomID == "" {
			http.Error(w, "missing room", http.StatusNotFound)
			return
		}
		sock.RegisterRoom(roomID)
		err := sock.Upgrade(w, r)
		if err != nil {
			log.Printf("Upgrade fail %v", err)
			http.Error(w, "connection fail", http.StatusBadRequest)
			return
		}
	})

	r.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		// do something, broadcast message
		sock.Broadcast(context.Background(), []byte("Hello world"))
	})

	log.Println("Service listening on port :8000")
	http.ListenAndServe(":8000", r)
}
