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
		conn := websocket.UpgradeHTTP(w, r)
		Serve(conn)
	}).GET()
	log.Fatal(http.ListenAndServe(":8080", m))
}

func Serve(conn *websocket.Conn) {
	for {
		var message string
		err := conn.ReadMsg(&message)
		if err != nil {
			break
		}
		conn.WriteMsg(strings.ToUpper(string(message)))
	}
	conn.Close()
}
