package ws

import (
	"net/http"
)

type WebSocket struct {
	Handler Handler
	W       http.ResponseWriter
	Request *http.Request
}
