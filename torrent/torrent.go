package torrent

import (
	"io"
	"crypto/rand"
)

type Torrent struct {
	PeerId 	 	[20]byte
	Announce    string
	PieceLength uint32
	Pieces 		[][20]byte
	Name  		string
	PieceCount	int
	InfoHash 	[20]byte
	Length 		int64
}

func New(r io.Reader) (*Torrent, error) {
	meta, err := NewInfo(r)
	if err != nil {
		return nil, err
	}
	var peerId [20]byte
	_, err = rand.Read(peerId[:])
	if err != nil {
		return nil, err
	}

	
	if err != nil {
		return nil, err
	}

	return &Torrent{
		PeerId: 	 peerId,
		Announce:    meta.Announce,
		PieceLength: meta.Info.PieceLength,
		Pieces: 	 splitPieces(meta.Info.Pieces),
		Name: 		 meta.Info.Name,
		InfoHash: 	 meta.InfoHash,
		PieceCount:  len([]byte(meta.Info.Pieces)) / 20,
		Length: 	 meta.Info.Length,
	}, nil
}

func splitPieces(piece string) [][20]byte {
	buff := []byte(piece)
	pieceCount := len(buff) / 20
	pieceBuffer := make([][20]byte, pieceCount)
	for i := 0; i < pieceCount; i++ {
		copy(pieceBuffer[i][:], buff[i*20: (i+1)*20])
	}
	return pieceBuffer
}