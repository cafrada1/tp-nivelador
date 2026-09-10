package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

const (
	payloadLengthSize = 4
	messageIDSize     = 4
	winnersLengthSize = 2

	messageTypeSize = 1
)

const (
	messageOpen    byte = 0x00
	messageData    byte = 0x01
	messageClose   byte = 0x02
	messageWinners byte = 0x03
	messageAck     byte = 0x04
)

const (
	betsLengthSize = 2
	nameLengthSize = 1
	documentSize   = 4
	birthdateSize  = 4
	betNumberSize  = 4
)

var (
	SizeOfUint8      = binary.Size(uint8(0))
	SizeOfUint16     = binary.Size(uint16(0))
	SizeOfUint32     = binary.Size(uint32(0))
	FatalToShortData = errors.New("fatal error: data too short")
	ValueOutOfRange  = errors.New("value out of range")
)

func DecodeUint8(data []byte, offset int) (int, error) {
	if len(data) < offset+SizeOfUint8 {
		return 0, ValueOutOfRange
	}
	return int(data[offset]), nil
}

func DecodeUint16(data []byte, offset int) (int, error) {
	endIndex := offset + SizeOfUint16
	if len(data) < endIndex {
		return 0, ValueOutOfRange
	}
	return int(binary.BigEndian.Uint16(data[offset:endIndex])), nil
}

func DecodeUint32(data []byte, offset int) (int, error) {
	endIndex := offset + SizeOfUint32
	if len(data) < endIndex {
		return 0, ValueOutOfRange
	}
	return int(binary.BigEndian.Uint32(data[offset:endIndex])), nil
}

func EncodeUint8(data *bytes.Buffer, value int) error {
	if value > math.MaxUint8 {
		return ValueOutOfRange
	}
	return binary.Write(data, binary.BigEndian, uint8(value))
}

func EncodeUint16(data *bytes.Buffer, value int) error {
	if value > math.MaxUint16 {
		return ValueOutOfRange
	}
	return binary.Write(data, binary.BigEndian, uint16(value))
}

func EncodeUint32(data *bytes.Buffer, value int) error {
	if value > math.MaxUint32 {
		return ValueOutOfRange
	}
	return binary.Write(data, binary.BigEndian, uint32(value))
}
