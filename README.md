# websocket
[![GoDoc](https://godoc.org/github.com/hslam/websocket?status.svg)](https://godoc.org/github.com/hslam/websocket)

Package websocket implements a client and server for the WebSocket protocol as specified in [RFC 6455](https://tools.ietf.org/html/rfc6455 "RFC 6455").

## Feature
* Upgrade HTTP/Conn
* TLS

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

**server.go**
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
	m.HandleFunc("/upper", func(w http.ResponseWriter, r *http.Request) {
		conn := websocket.UpgradeHTTP(w, r)
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

**client.go**
```go
package main

import (
	"fmt"
	"github.com/hslam/websocket"
	"time"
)

func main() {
	conn, err := websocket.Dial("tcp", "127.0.0.1:8080", "/upper", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for i := 0; i < 1; i++ {
		conn.SendMessage([]byte("Hello World"))
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
    var ws = null;
    var wsuri = "ws://127.0.0.1:8080/upper";
    ws = new WebSocket(wsuri);
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


### Authors
websocket was written by Meng Huang.


