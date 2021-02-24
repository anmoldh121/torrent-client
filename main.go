package main

import (
	"os"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/torrent-client/torrent"
	"github.com/torrent-client/p2p"
	"github.com/torrent-client/tracker"
)

const torrentPath = "/home/beast/youtube/torrent-client/torrent/test/ubuntu-14.04.6-server-ppc64el.iso.torrent"

func main() {
	f, err := os.Open(torrentPath)
	if err != nil {
		log.Error("Can not open file")
	}
	torrent, err := torrent.New(f)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(torrent.Name)
	tr, err := tracker.New(torrent, 6681)
	if err != nil {
		log.Error(err)
	}
	peers, err := tr.GetPeers()
	if err != nil {
		log.Error(err)
	}
	node := p2p.New(torrent, peers)
	progress := make(chan int)	
	buf, err := node.StartDownload(progress)
	if err != nil {
		log.Error("error in download ", err)
	}
	outFile, err := os.Create("/home/beast/youtube/torrent-client/torrent/test/" + torrent.Name)
	if err != nil {
		log.Error(err)
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		log.Error(err)
	}
}