package ws

import "net/http"

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
			socket.Recv()
		}
	}()

	for {
		select {
		case p := <-socket.OutCH:
			if err := socket.Handler.Send(p); err != nil {
				socket.Close()
			}
		case df := <-socket.InCH:
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
				socket.Close()
			}
		case <-socket.Closed:
			if s.OnClose != nil {
				s.OnClose(socket)
			}
			socket.Handler.Close()
			return
		}
	}
}
