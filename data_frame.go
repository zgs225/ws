package ws

import (
	"encoding/binary"
	"io"
)

type OpCode byte

type DataFrameHeader []byte

type MaskKey [4]byte

const (
	OpCodes_CONTINUATION OpCode = 0x0
	OpCodes_TEXT                = 0x1
	OpCodes_BINARY              = 0x2
	OpCodes_CLOSE               = 0x8
	OpCodes_PING                = 0x9
	OpCodes_PONG                = 0xA
)

const (
	DataFrame_BIT1 = 0x80
	DataFrame_BIT4 = 0x0F
	DataFrame_BIT7 = 0x7F
)

// DataFrameHeader methods

func parseDataFrameHeader(r io.ByteReader) (DataFrameHeader, error) {
	b0, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	b1, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	var remain int
	l := b1 & DataFrame_BIT7
	if l <= 125 {
		remain = 4
	} else if l == 126 {
		remain = 6
	} else {
		remain = 12
	}

	h := make([]byte, remain+2)
	h[0] = b0
	h[1] = b1
	for i := 0; i < remain; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		h[i+2] = b
	}

	return DataFrameHeader(h), nil
}

func (h DataFrameHeader) Length() uint64 {
	l := h[1] & DataFrame_BIT7
	if l <= 125 {
		return uint64(l)
	} else if l == 126 {
		return uint64(binary.BigEndian.Uint16(h[2:4]))
	} else {
		return binary.BigEndian.Uint64(h[2:10])
	}
}

func (h DataFrameHeader) IsMasked() bool {
	return h[1]&DataFrame_BIT1 != 0
}

func (h DataFrameHeader) MaskKey() MaskKey {
	l := len(h)
	b := [4]byte{h[l-4], h[l-3], h[l-2], h[l-1]}
	return MaskKey(b)
}

type DataFrame struct {
	Header  DataFrameHeader
	Payload []byte
}

func (df *DataFrame) GetPayload() []byte {
	b := make([]byte, len(df.Payload))
	if df.Header.IsMasked() {
		k := df.Header.MaskKey()
		for i := 0; i < len(df.Payload); i++ {
			b[i] = df.Payload[i] ^ k[i%4]
		}
	} else {
		copy(b, df.Payload)
	}
	return b
}
