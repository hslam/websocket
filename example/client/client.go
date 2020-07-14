package main

import (
	"fmt"
	"github.com/hslam/websocket"
	"time"
)

func main() {
	conn, err := websocket.Dial("127.0.0.1:8080", "/upper")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		conn.WriteMessage([]byte("Hello websocket"))
		var message string
		err := conn.ReadMessage(&message)
		if err != nil {
			break
		}
		fmt.Println(message)
		time.Sleep(time.Second)
	}
}
