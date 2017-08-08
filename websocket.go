package ws

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
)

const (
	WS_GUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

var (
	ErrHTTPVersion               = errors.New("HTTP version must be at least 1.1")
	ErrHostHeader                = errors.New("Host header required or error")
	ErrUpgradeHeader             = errors.New("Upgrade header required or error")
	ErrConnectionHeader          = errors.New("Connection header required or error")
	ErrSecWebSocketKeyHeader     = errors.New("Sec-WebSocket-Key header required or error")
	ErrOriginHeader              = errors.New("Origin header required or error")
	ErrSecWebSocketVersionHeader = errors.New("Sec-WebSocket-Version header must be 13")
)

type WebSocket struct {
	Handler Handler
	W       http.ResponseWriter
	Request *http.Request

	ID      uint
	OutCH   chan []byte
	InCH    chan *DataFrame
	Closed  bool
	CloseCH chan struct{}
}

func NewWebSocket(w http.ResponseWriter, request *http.Request, before func(*WebSocket) (error, int)) (ws *WebSocket, err error) {
	ws = &WebSocket{
		W:       w,
		Request: request,
		ID:      NextGlobalID(),
		OutCH:   make(chan []byte),
		InCH:    make(chan *DataFrame),
		CloseCH: make(chan struct{}),
	}

	if before != nil {
		var code int
		if err, code = before(ws); err != nil {
			http.Error(w, err.Error(), code)
			return
		}
	}

	if err = ws.Handshake(); err != nil {
		return
	}

	var hdr Handler
	if hdr, err = newHandler(w); err != nil {
		return
	}
	ws.Handler = hdr

	return
}

func (ws *WebSocket) Send(p []byte) {
	ws.OutCH <- p
}

func (ws *WebSocket) Recv() {
	df, err := ws.Handler.Recv()
	if err != nil {
		ws.Close()
	}
	ws.InCH <- df
}

func (ws *WebSocket) Close() {
	ws.CloseCH <- struct{}{}
}

func (ws *WebSocket) Handshake() error {
	if err := ws.handshakeCheck(); err != nil {
		http.Error(ws.W, err.Error(), http.StatusBadRequest)
		return err
	}

	k := ws.Request.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(k))
	h.Write([]byte(WS_GUID))
	sec := base64.StdEncoding.EncodeToString(h.Sum(nil)[:])

	ws.W.Header().Set("Sec-WebSocket-Accept", sec)
	ws.W.Header().Set("Upgrade", "websocket")
	ws.W.Header().Set("Connection", "Upgrade")
	ws.W.WriteHeader(http.StatusSwitchingProtocols)

	return nil
}

func (ws *WebSocket) handshakeCheck() error {
	// Check HTTP version
	if ws.Request.ProtoMajor < 1 || (ws.Request.ProtoMajor == 1 && ws.Request.ProtoMinor < 1) {
		return ErrHTTPVersion
	}

	// Check Upgrade header
	upgrade := ws.Request.Header.Get("Upgrade")
	if len(upgrade) != 9 || upgrade != "websocket" {
		return ErrUpgradeHeader
	}

	// Check Connection header
	conn := ws.Request.Header.Get("Connection")
	if len(conn) != 7 || conn != "Upgrade" {
		return ErrConnectionHeader
	}

	// Check Sec-WebSocket-Key header
	key := ws.Request.Header.Get("Sec-WebSocket-Key")
	if b, err := base64.StdEncoding.DecodeString(key); err != nil {
		return err
	} else {
		if len(b) != 16 {
			return ErrSecWebSocketKeyHeader
		}
	}

	// Check Origin header
	origin := ws.Request.Header.Get("Origin")
	if len(origin) == 0 {
		return ErrOriginHeader
	}

	// Check Sec-WebSocket-Version header
	if ws.Request.Header.Get("Sec-WebSocket-Version") != "13" {
		return ErrSecWebSocketVersionHeader
	}

	return nil
}
