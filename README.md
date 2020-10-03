# websocket
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hslam/websocket)](https://pkg.go.dev/github.com/hslam/websocket)
[![Build Status](https://travis-ci.org/hslam/websocket.svg?branch=master)](https://travis-ci.org/hslam/websocket)
[![Go Report Card](https://goreportcard.com/badge/github.com/hslam/websocket)](https://goreportcard.com/report/github.com/hslam/websocket)
[![LICENSE](https://img.shields.io/github/license/hslam/websocket.svg?style=flat-square)](https://github.com/hslam/websocket/blob/master/LICENSE)

Package websocket implements a client and server for the WebSocket protocol as specified in [RFC 6455](https://tools.ietf.org/html/rfc6455 "RFC 6455").

## Feature
* Upgrade HTTP / Conn
* TLS

## [Benchmark](https://github.com/hslam/websocket-benchmark "websocket-benchmark")

##### Websocket QPS

<img src="https://raw.githubusercontent.com/hslam/websocket/master/websocket-qps.png"  alt="websocket" align=center>


## Get started

### Install
```
go get github.com/hslam/websocket
```
### Import
```
import "github.com/hslam/websocket"
```
### Usage
#### Example

server.go
```go
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
```

server_poll.go
```go
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
```

client.go
```go
package main

import (
	"fmt"
	"github.com/hslam/websocket"
	"time"
)

func main() {
	conn, err := websocket.Dial("tcp", "127.0.0.1:8080", "/", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for i := 0; i < 1; i++ {
		conn.SendMessage("Hello World")
		var message string
		err := conn.ReceiveMessage(&message)
		if err != nil {
			break
		}
		fmt.Println(message)
		time.Sleep(time.Second)
	}
}
```

**Output**
```
HELLO WORLD
```

**client.html**
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8"/>
    <title>Websocket</title>
</head>
<body>
<h1>UPPER</h1>
<form><p>string: <input id="content" type="text" placeholder="input string"></p></form>
<label id="result">result：</label><br><br>
<button onclick="send()">upper</button>
<script type="text/javascript">
    var wsuri = "ws://127.0.0.1:8080/";
    var ws = new WebSocket(wsuri);
    ws.onmessage = function(e) {
        var result = document.getElementById('result');
        result.innerHTML = "result：" + e.data;
    }
    function send() {
        var msg = document.getElementById('content').value;
        ws.send(msg);
    }
</script>
</body>
</html>
```

### License
This package is licensed under a MIT license (Copyright (c) 2020 Meng Huang)


### Author
websocket was written by Meng Huang.


