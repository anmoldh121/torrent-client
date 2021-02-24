package peers

import (
	"net"
	"time"
	"bytes"
	"fmt"
	"io"

	message "github.com/torrent-client/messages"
	"github.com/torrent-client/handshake"
)

type RemotePeer struct {
	Conn     net.Conn
	PeerId   [20]byte
	Choked   bool
	Peer 	 Peer
	InfoHash [20]byte
	Bitfield message.Bitfield
}

func makeHandshake(conn net.Conn, infoHash, peerId [20]byte) (*handshake.Handshake, error) {
	req := handshake.New(infoHash, peerId)
	conn.SetDeadline(time.Now().Add(3*time.Second))
	defer conn.SetDeadline(time.Time{})
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn, conn)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("Connection closed by remote peer")
		}
		fmt.Println("Error in reading")
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("Infohash does not collide")
	}

	return res, nil
}


func GetBitfield(conn net.Conn) (message.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5*time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}	

	if msg == nil {
		return nil, fmt.Errorf("Expected bitfield but got %s", msg)
	}

	if msg.ID != message.MBitfield {
		return nil, fmt.Errorf("Expected bitfield but got %d", msg.ID)
	}
	return msg.Payload, nil
}

func NewRemotePeer(infoHash [20]byte, peer Peer, peerId [20]byte) (*RemotePeer, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}	
	_, err = makeHandshake(conn, infoHash, peerId)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := GetBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &RemotePeer{
		Conn: conn,
		Bitfield: bf,
		PeerId: peerId,
		Choked: true,
		InfoHash: infoHash,
		Peer: peer,
	}, nil
}

func (rp *RemotePeer) SendUnchoke() error {
	msg := message.Message{ID: message.MUnchoke}
	_, err := rp.Conn.Write(msg.Serialize())
	return err
}

func (rp *RemotePeer) SendInterested() error {
	msg := message.Message{ID: message.MInterested}
	_, err := rp.Conn.Write(msg.Serialize())
	return err
}

func (rp *RemotePeer) SendRequest(index, begin, length int) error {
	msg := message.NewRequest(index, begin, length)
	_, err := rp.Conn.Write(msg.Serialize())
	return err
}

func (c *RemotePeer) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}