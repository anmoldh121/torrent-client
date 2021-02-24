package p2p

import (
	"fmt"
	"sync"
	"time"
	"crypto/sha1"
	"bytes"
	"runtime"
	"github.com/torrent-client/torrent"
	"github.com/torrent-client/peers"
	message "github.com/torrent-client/messages"
	"github.com/torrent-client/handshake"
	log "github.com/sirupsen/logrus"
)

const MaxBacklog = 5

const MaxBlockSize = 16384

type Node struct {
	Torrent *torrent.Torrent
	Peers 	[]peers.Peer
	wg 		*sync.WaitGroup
}

type Progress struct {
	index 	   int 
	client 	   *peers.RemotePeer
	buff 	   []byte
	downloaded int
	requested  int
	backlog    int 
}

type pieceResult struct {
	index int
	buf   []byte
}

type RemotePeerInfo struct {
	index  int
	hash   [20]byte
	length int
}

func New(tr *torrent.Torrent, peerList []peers.Peer) *Node {
	var wg sync.WaitGroup
	return &Node{
		Torrent: tr,
		Peers: 	 peerList,
		wg	 :	 &wg,
	}
}

func (node *Node) CalculatePieceBound(index int) (int, int) {
	begin := index*int(node.Torrent.PieceLength)
	end := begin + int(node.Torrent.PieceLength)
	return begin, end
}

func (node *Node) CalculatePieceSize(index int) int {
	begin, end := node.CalculatePieceBound(index)
	return end - begin
}

func (state *Progress) ReadMessage() error {
	msg, err := message.Read(state.client.Conn)
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MUnchoke:
		state.client.Choked = false
	case message.MChoke:
		state.client.Choked = true
	case message.MHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MPiece:
		n, err := message.ParsePiece(state.index, state.buff, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}

	return nil
}

func attemptDownload(remotePeer *peers.RemotePeer, p *RemotePeerInfo) ([]byte, error) {
	state := &Progress{
		index: p.index,
		client: remotePeer,
		buff: make([]byte, p.length),
		downloaded: 0,
	}

	remotePeer.Conn.SetDeadline(time.Now().Add(30*time.Second))
	defer remotePeer.Conn.SetDeadline(time.Time{})

	for state.downloaded < p.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < p.length {
				blockSize := MaxBlockSize
				if p.length-state.requested < blockSize {
					blockSize = p.length - state.requested
				}

				err := remotePeer.SendRequest(p.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}
		err := state.ReadMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buff, nil
}

func checkIntegrity(pw *RemotePeerInfo, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}

func (node *Node) StartDownloader(peer peers.Peer, queue chan *RemotePeerInfo, results chan *pieceResult) {
	node.wg.Add(1)
	_ = handshake.New(node.Torrent.InfoHash, node.Torrent.PeerId)
	remotePeer, err := peers.NewRemotePeer(node.Torrent.InfoHash, peer, node.Torrent.PeerId)
	node.wg.Done()
	if err != nil {
		log.Error("Error in hanshake ", err)
		return
	}
	log.Info("Handshake complete")
	defer remotePeer.Conn.Close()

	remotePeer.SendUnchoke()
	remotePeer.SendInterested()

	for q := range queue {
		if !remotePeer.Bitfield.HasPiece(q.index) {
			queue <- q
			continue
		}

		buf, err := attemptDownload(remotePeer, q)
		if err != nil {
			fmt.Println("Exiting")
			queue <- q
			return
		}
		err = checkIntegrity(q, buf)
		if err != nil {
			fmt.Println("Integrity check failed")
			queue <- q
			continue
		}

		err = remotePeer.SendHave(q.index)
		if err != nil {
			fmt.Println("Failed to send have")
			continue
		}
		results <- &pieceResult{q.index, buf}
	}
}

func (node *Node) StartDownload(progress chan int) ([]byte, error) {
	fmt.Println(node.Peers)
	queue := make(chan *RemotePeerInfo, len(node.Torrent.Pieces))
	results := make(chan *pieceResult)
	for i, piece := range node.Torrent.Pieces {
		queue <- &RemotePeerInfo{
			index: i,
			hash: piece,
			length: node.CalculatePieceSize(i),
		}
	}

	for _, peer := range node.Peers {
		go node.StartDownloader(peer, queue, results)
		node.wg.Wait()
	}

	buf := make([]byte, node.Torrent.Length)
	donePieces := 0
	for donePieces < len(node.Torrent.Pieces) {
		res := <-results
		begin, end := node.CalculatePieceBound(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(node.Torrent.Pieces)) * 100
		numWorkers := runtime.NumGoroutine() - 1 
		log.Infof("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(queue)

	return buf, nil
}