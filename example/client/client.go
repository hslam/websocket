package main

import (
	"bufio"
	"fmt"
	"github.com/hslam/websocket"
	"io"
	"time"
)

func main() {
	conn, err := websocket.Dial("127.0.0.1:8080", "/upper")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for i := 0; i < 3; i++ {
		conn.Write([]byte("Hello websocket\n"))
		message, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		fmt.Print(message)
		time.Sleep(time.Second)
	}
}
