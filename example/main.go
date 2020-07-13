package main

import (
	"bufio"
	"github.com/hslam/mux"
	"github.com/hslam/websocket"
	"io"
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
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		conn.Write([]byte(strings.ToUpper(string(message))))
	}
}
