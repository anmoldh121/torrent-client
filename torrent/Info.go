package torrent

import (
	"io"
	"crypto/sha1"

	"github.com/zeebo/bencode"
)

type Info struct {
	PieceLength uint32 	`bencode:"piece length"`
	Pieces 		string	`bencode:"pieces"`
	// Private		string 	`bencode:"private"`
	Name		string	`bencode:"name"`
	Length 		int64 	`bencode:"length"`
}

type MetaInfo struct {
	Info 	 Info 
	Announce string
	InfoHash [20]byte 
}

func NewInfo(r io.Reader) (MetaInfo, error) {
	var s struct{
		Info 	 bencode.RawMessage `bencode:"info"`
		Announce bencode.RawMessage `bencode:"announce"`
	}
	meta := MetaInfo{}
	err := bencode.NewDecoder(r).Decode(&s)
	if err != nil {
		return MetaInfo{}, nil
	}
	err = bencode.DecodeBytes(s.Info, &meta.Info)
	if err != nil {
		return MetaInfo{}, nil
	}
	hash := sha1.New()
	_, _ = hash.Write(s.Info)
	copy(meta.InfoHash[:], hash.Sum(nil))
	err = bencode.DecodeBytes(s.Announce, &meta.Announce)
	if err != nil {
		return MetaInfo{}, err
	}
	return meta, err
}
