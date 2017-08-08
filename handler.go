package ws

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var (
	ErrHijacker = errors.New("Can't convert to http.Hijacker")
)

type Handler interface {
	Recv() (*DataFrame, error)

	Send(p []byte) error

	Ping() error

	Pong() error

	Close() error
}

type handler struct {
	Conn  net.Conn
	Bufrw *bufio.ReadWriter
}

func newHandler(w http.ResponseWriter) (Handler, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrHijacker
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}
	return &handler{Conn: conn, Bufrw: bufrw}, nil
}

func (h *handler) Recv() (*DataFrame, error) {
	frame := new(DataFrame)
	header, err := parseDataFrameHeader(h.Bufrw)
	if err != nil {
		return nil, err
	}
	frame.Header = header
	frame.Payload = make([]byte, frame.Header.Length())
	_, err := h.Bufrw.Read(frame.Payload)
	if err != nil {
		return nil, err
	}
	return frame, nil
}

func (h *handler) Send(p []byte) error {
	df := DataFrame{}
	if _, err := df.Write(p); err != nil {
		return err
	}
	return h.send(df)
}

func (h *handler) send(df *DataFrame) error {
	if _, err := h.Bufrw.Write(df.Bytes()); err != nil {
		return err
	}
	return h.Bufrw.Flush()
}

func (h *handler) Close() error {
	hdr := DataFrameHeader{OpCodes_CLOSE | DataFrame_BIT1, 0}
	frm := &DataFrame{Header: hdr}
	if err := h.send(frm); err != nil {
		return err
	}
	return h.Conn.Close()
}

func (h *handler) Ping() error {
	hdr := DataFrameHeader{OpCodes_PING | DataFrame_BIT1, 0}
	frm := &DataFrame{Header: hdr}
	return h.send(frm)
}

func (h *handler) Pong() error {
	hdr := DataFrameHeader{OpCodes_PONG | DataFrame_BIT1, 0}
	frm := &DataFrame{Header: hdr}
	return h.send(frm)
}
