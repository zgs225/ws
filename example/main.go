package main

import (
	"log"
	"net/http"

	"github.com/zgs225/ws"
)

func main() {
	server := new(ws.Server)

	log.Fatal(http.ListenAndServe(":8080", server))
}
