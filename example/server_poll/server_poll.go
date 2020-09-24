package main

import (
	"github.com/hslam/netpoll"
	"github.com/hslam/websocket"
	"net"
	"strings"
)

func main() {
	var handler = &netpoll.ConnHandler{}
	handler.SetUpgrade(func(conn net.Conn) (netpoll.Context, error) {
		return websocket.Upgrade(conn, nil)
	})
	handler.SetServe(func(context netpoll.Context) error {
		ws := context.(*websocket.Conn)
		var message string
		err := ws.ReceiveMessage(&message)
		if err != nil {
			return err
		}
		return ws.SendMessage(strings.ToUpper(string(message)))
	})
	if err := netpoll.ListenAndServe("tcp", ":8080", handler); err != nil {
		panic(err)
	}
}
