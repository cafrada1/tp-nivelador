package protocol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type Decoder interface {
	ReceiveWinners() (domain.Bets, error)
}

type decoder struct {
	sock io.Reader
}

func NewDecoder(sock io.Reader) *decoder {
	return &decoder{sock}
}

func (d *decoder) ReceiveWinners() (domain.Bets, error) {
	payloadLength, err := d.readPayloadLength()
	if err != nil {
		return nil, err
	}
	data, err := safe_socket.RecvAll(d.sock, payloadLength)
	if err != nil {
		return nil, err
	}

	return decodeWinners(data)
}

func (d *decoder) readPayloadLength() (int, error) {
	payloadLength, err := safe_socket.RecvAll(d.sock, payloadLengthSize)
	if err != nil {
		return 0, err
	}

	return int(binary.BigEndian.Uint32(payloadLength)), nil
}

func decodeWinners(data []byte) (domain.Bets, error) {
	length, err := decodeBetsLength(data[:winnersLengthSize])
	if err != nil {
		return nil, err
	}
	offset := winnersLengthSize

	winners := make(domain.Bets, length)
	var winner domain.Bet
	for i := range length {
		winner, offset = decodeWinner(data, offset)
		winners[i] = winner
	}
	return winners, nil
}

func decodeBetsLength(data []byte) (int, error) {
	return int(binary.BigEndian.Uint32(data)), nil
}

func decodeWinner(data []byte, offset int) (domain.Bet, int) {
	firstName, offset := decodeName(data, offset)
	lastName, offset := decodeName(data, offset)
	document, offset := decodeDocument(data, offset)
	birthdate, offset := decodeBirthdate(data, offset)
	betNumber, offset := decodeBetNumber(data, offset)
	return domain.Bet{
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    betNumber,
	}, offset
}

func decodeBetNumber(data []byte, offset int) (int, int) {
	document := int(binary.BigEndian.Uint16(data[offset : offset+betNumberSize]))
	offset += betNumberSize
	return document, offset
}

func decodeBirthdate(data []byte, offset int) (string, int) {
	birthdate := int(binary.BigEndian.Uint32(data[offset : offset+birthdateSize]))
	offset += birthdateSize
	return fmt.Sprintf("%04d-%02d-%02d", birthdate/10000, (birthdate%10000)/100, birthdate%100), offset
}

func decodeDocument(data []byte, offset int) (int, int) {
	document := int(binary.BigEndian.Uint32(data[offset : offset+documentSize]))
	offset += documentSize
	return document, offset
}

func decodeName(data []byte, offset int) (string, int) {
	length := int(data[offset : offset+nameLengthSize][0])
	offset += nameLengthSize
	name := string(data[offset : offset+length])
	offset += length
	return name, offset
}
