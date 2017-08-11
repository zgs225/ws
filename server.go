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
	OnError         func(*WebSocket, error)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	socket, err := NewWebSocket(w, request, s.BeforeHandshake)

	if err != nil && s.OnError != nil {
		s.OnError(socket, err)
		return
	}

	if s.OnOpen != nil {
		if err := s.OnOpen(socket); err != nil {
			socket.Close()
		}
	}

	go func() {
		for {
			if socket.Status == WebSocketStatus_CLOSING || socket.Status == WebSocketStatus_CLOSED {
				return
			}
			socket.Recv()
		}
	}()

	go func() {
		ticker := time.Tick(time.Second)
		for range ticker {
			if socket.Status == WebSocketStatus_CLOSING || socket.Status == WebSocketStatus_CLOSED {
				return
			}
			if err := socket.Handler.Ping(); err != nil {
				socket.Close()
				return
			}
		}
	}()

	for {
		if socket.Status == WebSocketStatus_CLOSED {
			return
		}
		select {
		case <-socket.CloseCH:
			s.close(socket)
			return
		case p := <-socket.OutCH:
			if socket.Status == WebSocketStatus_CLOSING || socket.Status == WebSocketStatus_CLOSED {
				continue
			}
			if err := socket.Handler.Send(p); err != nil {
				socket.Close()
			}
		case df := <-socket.InCH:
			if socket.Status == WebSocketStatus_CLOSING || socket.Status == WebSocketStatus_CLOSED {
				continue
			}
			switch df.Header.GetOpCode() {
			case OpCodes_CONTINUATION:
				// TODO
			case OpCodes_BINARY:
				// TODO
			case OpCodes_TEXT:
				if s.OnMessage != nil {
					if err := s.OnMessage(df, socket); err != nil {
						socket.Close()
					}
				}
			case OpCodes_PING:
				socket.Handler.Pong()
			case OpCodes_CLOSE:
				s.close(socket)
			}
		}
	}
}

func (s *Server) close(socket *WebSocket) {
	socket.Status = WebSocketStatus_CLOSING
	if s.OnClose != nil {
		s.OnClose(socket)
	}
	socket.Handler.Close()
	socket.Status = WebSocketStatus_CLOSED
}
