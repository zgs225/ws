WS
===

Simple websocket server for golang.

### Usage

``` go
func main() {
	server := &ws.Server{
		OnMessage: func(df *ws.DataFrame, socket *ws.WebSocket) error {
			log.Println("Dataframe received.")
			log.Println("Payload: ", string(df.GetPayload()))
			return nil
		},
		OnError: func(socket *ws.WebSocket, err error) {
			log.Println(err)
		},
		OnClose: func(socket *ws.WebSocket) error {
			log.Println("socket closed")
			return nil
		},
	}

	log.Fatal(http.ListenAndServe(":8080", server))
}
```

### Callbacks

+ `BeforeHandshake: func(*WebSocket) (error, int)`
+ `OnOpen: func(*WebSocket) error`
+ `OnMessage: func(*DataFrame, *WebSocket) error`
+ `OnClose: func(*WebSocket) error`
+ `OnError: func(*WebSocket, error)`
