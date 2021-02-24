package handshake

import (
	"io"
	"fmt"
	"net"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerId   [20]byte
}

func New(infoHash [20]byte, peerId [20]byte) *Handshake {
	return &Handshake{
		Pstr: 	  "BitTorrent protocol",
		InfoHash: infoHash,
		PeerId:   peerId,
	}
}

func (h *Handshake) Serialize() []byte {
	buff := make([]byte, 49+len(h.Pstr))
	buff[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buff[curr:], h.Pstr)
	curr += copy(buff[curr:], make([]byte, 8))
	curr += copy(buff[curr:], h.InfoHash[:])
	curr += copy(buff[curr:], h.PeerId[:])
	return buff
}

func Read(r io.Reader, conn net.Conn) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		return nil, fmt.Errorf("pstrlen can not be zero")
	}

	handshakeBuf := make([]byte, 48+pstrlen)
	
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}
	var infoHash, peerId [20]byte
	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+28])
	copy(peerId[:], handshakeBuf[pstrlen+28:])
	return &Handshake{
		Pstr:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerId:	  peerId,
	}, nil
}