package main

import (
	"github.com/hslam/mux"
	"github.com/hslam/websocket"
	"log"
	"net/http"
	"strings"
)

func main() {
	m := mux.New()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.UpgradeHTTP(w, r)
		if err != nil {
			return
		}
		for {
			var message string
			err := conn.ReceiveMessage(&message)
			if err != nil {
				break
			}
			conn.SendMessage(strings.ToUpper(string(message)))
		}
		conn.Close()
	}).GET()
	log.Fatal(http.ListenAndServe(":8080", m))
}
