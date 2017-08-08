package main

import (
	"log"
	"net/http"

	"github.com/zgs225/ws"
)

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
	}

	log.Fatal(http.ListenAndServe(":8080", server))
}
