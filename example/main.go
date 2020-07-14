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
	m.HandleFunc("/upper", func(w http.ResponseWriter, r *http.Request) {
		conn := websocket.Accept(w, r)
		ServeConn(conn)
	}).GET()
	log.Fatal(http.ListenAndServe(":8080", m))
}

func ServeConn(conn *websocket.Conn) {
	for {
		var message string
		err := conn.ReadMessage(&message)
		if err != nil {
			break
		}
		conn.WriteMessage(strings.ToUpper(string(message)))
	}
}
