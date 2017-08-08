package ws

import (
	"net/http"
	"time"
)

type Server struct {
	BeforeHandshake func(*WebSocket) (error, int)
	OnOpen          func(*WebSocket) error
	OnClose         func(*WebSocket) error
	OnMessage       func(*DataFrame, *WebSocket) error
}

func (s *Server) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	ws, err := NewWebSocket(w, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.BeforeHandshake != nil {
		if err, code := s.BeforeHandshake(ws); err != nil {
			http.Error(w, err.Error(), code)
			return
		}
	}

	if err := ws.Handshake(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.OnOpen != nil {
		if err := s.OnOpen(ws); err != nil {
			ws.Close()
		}
	}

	go func() {
		for {
			ws.Recv()
		}
	}()

	for {
		select {
		case df := <-ws.OutCH:
			if err := ws.Handler.Send(df); err != nil {
				ws.Close()
			}
		case df := <-ws.InCH:
			switch df.Header.GetOpCode() {
			case OpCodes_CONTINUATION:
				// TODO
			case OpCodes_BINARY:
				// TODO
			case OpCodes_TEXT:
				if s.OnMessage != nil {
					if err := s.OnMessage(df, ws); err != nil {
						ws.Close()
					}
				}
			case OpCodes_PING:
				ws.Handler.Pong()
			case OpCodes_CLOSE:
				ws.Close()
			}
		case <-ws.Closed:
			if s.OnClose != nil {
				s.OnClose(ws)
			}
			ws.Handler.Close()
		}
	}
}
