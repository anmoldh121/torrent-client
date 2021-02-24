package message

import (
	"io"
	"encoding/binary"
	"fmt"
)

type MessageID uint8

const (
	MChoke 		   MessageID = 0
	MUnchoke 	   MessageID = 1
	MInterested    MessageID = 2
	MNotInterested MessageID = 3
	MHave 		   MessageID = 4
	MBitfield 	   MessageID = 5
	MRequest 	   MessageID = 6
	MPiece		   MessageID = 7
	MCancel		   MessageID = 8
	MPort		   MessageID = 9
)

type Message struct {
	ID 		MessageID
	Payload []byte
}

type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8 
	if byteIndex < 0 || byteIndex >= len(b) {
		return false
	}	
	return b[byteIndex]>>(7-offset)&1 != 0
}

func (b Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8 
	if byteIndex < 0 || byteIndex >= len(b) {
		return
	}	
	b[byteIndex] |= 1<<(7-offset)
}


func Read(r io.Reader) (*Message, error) {
	lengthBuff := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuff)
	length := binary.BigEndian.Uint32(lengthBuff)

	if length == 0 {
		return nil, nil
	}

	messageBuff := make([]byte, length)
	_, err = io.ReadFull(r, messageBuff)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID: MessageID(messageBuff[0]),
		Payload: messageBuff[1:],
	}, nil
}

func (msg *Message) Serialize() []byte {
	if msg == nil {
		return make([]byte, 4)
	}
	length := uint32(len(msg.Payload) + 1)
	buff := make([]byte, length + 4)
	binary.BigEndian.PutUint32(buff[0:4], length)
	buff[4] = byte(msg.ID)
	copy(buff[5:], msg.Payload)
	return buff
}

func NewRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MRequest, Payload: payload}
}

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MHave, Payload: payload}
}

func ParseHave(msg *Message) (int, error) {
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected payload length 4")
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MPiece {
		return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", MPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Payload too short. %d < 8", len(msg.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}