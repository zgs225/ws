package ws

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

const (
	WS_GUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

type Server struct{}

func (s *Server) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	log.Println("ws connecting...")
	s.handshake(w, request)

	hj := w.(http.Hijacker)
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		conn.Close()
		log.Println(err)
	}

	for {
		frame := new(DataFrame)
		header, err := parseDataFrameHeader(bufrw)
		if err != nil {
			conn.Close()
			log.Println(err)
		}
		frame.Header = header
		frame.Payload = make([]byte, frame.Header.Length())
		bufrw.Read(frame.Payload)

		fmt.Printf("Payload length: %d\n", frame.Header.Length())
		fmt.Printf("Masked: %v\n", frame.Header.IsMasked())
		fmt.Printf("%0b\n", frame.Header)
		fmt.Printf("%s\n", frame.GetPayload())
	}
}

func (s *Server) handshake(w http.ResponseWriter, request *http.Request) {
	k := request.Header.Get("Sec-WebSocket-Key")
	h := sha1.New()
	h.Write([]byte(k))
	h.Write([]byte(WS_GUID))
	sec := base64.StdEncoding.EncodeToString(h.Sum(nil)[:])

	w.Header().Set("Sec-WebSocket-Accept", sec)
	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")
	w.WriteHeader(http.StatusSwitchingProtocols)
}
