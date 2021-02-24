package tracker

import (
	"net/url"
	"net/http"
	"time"
	"strconv"
	"github.com/torrent-client/torrent"
	"github.com/torrent-client/peers"
	"github.com/zeebo/bencode"
)

type Tracker struct {
	RawUrl *url.URL
	Params url.Values
}

type TrackerResp struct {
	TrackerId string `bencode:"tracker id"`
	Peers 	  string `bencode:"peers"`
	Interval  int 	 `bencode:"interval"`
}

func New(tr *torrent.Torrent, port uint16) (*Tracker, error) {
	base, err := url.Parse(tr.Announce)
	if err != nil {
		return nil, err
	}
	params := url.Values{
		"info_hash":  []string{string(tr.InfoHash[:])},
		"peer_id": 	  []string{string(tr.PeerId[:])},
		"port": 	  []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact": 	  []string{"1"},
		"left": 	  []string{strconv.Itoa(int(tr.Length))},
	}

	return &Tracker{
		RawUrl: base,
		Params: params,
	},nil
}

func (t *Tracker) URL() string {
	t.RawUrl.RawQuery = t.Params.Encode()
	return t.RawUrl.String()
}

func (t *Tracker) GetPeers() ([]peers.Peer, error) {
	base := t.URL()
	client := &http.Client{Timeout: 15*time.Second}
	resp, err := client.Get(base)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	trackerResp := TrackerResp{}
	err = bencode.NewDecoder(resp.Body).Decode(&trackerResp)	
	if err != nil {
		return nil, err
	}
	return peers.Unmarshal([]byte(trackerResp.Peers)), nil
}
