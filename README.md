# websocket
[RFC 6455](https://tools.ietf.org/html/rfc6455 "RFC 6455") - The WebSocket Protocol.
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
		conn := websocket.Upgrade(w, r)
		ServeConn(conn)
	}).GET()
	log.Fatal(http.ListenAndServe(":8080", m))
}

func ServeConn(conn *websocket.Conn) {
	for {
		var message string
		err := conn.ReadMsg(&message)
		if err != nil {
			break
		}
		conn.WriteMsg(strings.ToUpper(string(message)))
	}
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
	conn, err := websocket.Dial("127.0.0.1:8080", "/upper")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for i := 0; i < 3; i++ {
		conn.WriteMsg([]byte("Hello websocket"))
		var message string
		err := conn.ReadMsg(&message)
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
HELLO WEBSOCKET
HELLO WEBSOCKET
HELLO WEBSOCKET
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


