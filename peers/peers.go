package peers

import (
	"net"
	"fmt"
	"encoding/binary"
)

type Peer struct {
	IP	 net.IP
	Port uint16
}

func Unmarshal(resp []byte) []Peer {
	const ipSize = 4
	const portSize = 2
	const peerSize = 6
	fmt.Println(len(resp))
	numPeers := len(resp) / peerSize
	peers := make([]Peer, numPeers)

	for i := 0; i<numPeers; i++ {
		peers[i].IP = net.IP(resp[i*peerSize:i*peerSize+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(resp[i*peerSize+4:i*peerSize+6]))
	}
	return peers
}

func (peer *Peer) String() string {
	return peer.IP.String() + ":" + fmt.Sprint(peer.Port)
}

