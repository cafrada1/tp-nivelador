package protocol

import (
	"fmt"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type WinnersMessage struct {
	Id      int
	Winners domain.Bets
}

type Receiver interface {
	ReceiveWinners() (WinnersMessage, error)
	ReceiveAck() (int, error)
}

type receiver struct {
	sock io.Reader
}

type message struct {
	typeByte byte
	id       int
	data     []byte
}

func NewReceiver(sock io.Reader) *receiver {
	return &receiver{sock}
}

func (d *receiver) receive() (message, error) {
	header, err := safe_socket.RecvAll(d.sock, payloadLengthSize)
	if err != nil {
		return message{}, err
	}

	payloadLength, err := DecodeUint32(header, 0)
	if err != nil {
		return message{}, err
	}

	payload, err := safe_socket.RecvAll(d.sock, payloadLength)
	if err != nil {
		return message{}, err
	}

	if len(payload) < messageIDSize+messageTypeSize {
		return message{}, fmt.Errorf("fatal error: payload too short")
	}

	return d.decodePayload(payload)
}

func (d *receiver) decodePayload(payload []byte) (message, error) {
	offset := 0
	messageID, err := DecodeUint32(payload, offset)
	if err != nil {
		return message{}, err
	}
	offset += SizeOfUint32

	typeByte := payload[offset]
	offset += messageTypeSize

	data := payload[offset:]

	msg := message{
		typeByte: typeByte,
		id:       messageID,
		data:     data,
	}
	return msg, nil
}

func (d *receiver) ReceiveAck() (int, error) {
	msg, err := d.receive()
	if err != nil {
		return 0, err
	}
	if msg.typeByte != messageAck {
		return 0, fmt.Errorf("expected ACK message, got type %d", msg.typeByte)
	}
	return msg.id, nil
}

func (d *receiver) ReceiveWinners() (WinnersMessage, error) {
	msg, err := d.receive()
	if err != nil {
		return WinnersMessage{}, err
	}
	if msg.typeByte != messageWinners {
		return WinnersMessage{}, fmt.Errorf("expected WINNERS message, got type %d", msg.typeByte)
	}

	winners, err := decodeWinners(msg.data)
	if err != nil {
		return WinnersMessage{}, err
	}
	return WinnersMessage{Id: msg.id, Winners: winners}, nil
}

func decodeWinners(data []byte) (domain.Bets, error) {
	length, err := DecodeUint32(data, 0)
	if err != nil {
		return nil, err
	}
	offset := winnersLengthSize

	winners := make(domain.Bets, length)
	var winner domain.Bet
	for i := range length {
		winner, offset, err = decodeWinner(data, offset)
		if err != nil {
			return nil, err
		}
		winners[i] = winner
	}
	return winners, nil
}

func decodeWinner(data []byte, offset int) (domain.Bet, int, error) {
	firstName, offset, err := decodeName(data, offset)
	if err != nil {
		return domain.Bet{}, 0, err
	}
	lastName, offset, err := decodeName(data, offset)
	if err != nil {
		return domain.Bet{}, 0, err
	}
	document, offset, err := decodeDocument(data, offset)
	if err != nil {
		return domain.Bet{}, 0, err
	}
	birthdate, offset, err := decodeBirthdate(data, offset)
	if err != nil {
		return domain.Bet{}, 0, err
	}
	betNumber, offset, err := decodeBetNumber(data, offset)
	if err != nil {
		return domain.Bet{}, 0, err
	}

	bet := domain.Bet{
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    betNumber,
	}
	return bet, offset, nil
}

func decodeBetNumber(data []byte, offset int) (int, int, error) {
	number, err := DecodeUint32(data, offset)
	if err != nil {
		return 0, 0, err
	}
	offset += SizeOfUint32
	return number, offset, nil
}

func decodeBirthdate(data []byte, offset int) (string, int, error) {
	const formatDate = "%04d-%02d-%02d"
	birthdate, err := DecodeUint32(data, offset)
	if err != nil {
		return "", 0, err
	}
	offset += SizeOfUint32
	birthdateStr := fmt.Sprintf(formatDate, birthdate/10000, (birthdate%10000)/100, birthdate%100)
	return birthdateStr, offset, nil
}

func decodeDocument(data []byte, offset int) (int, int, error) {
	document, err := DecodeUint32(data, offset)
	if err != nil {
		return 0, 0, err
	}
	offset += SizeOfUint32
	return document, offset, nil
}

func decodeName(data []byte, offset int) (string, int, error) {
	length, err := DecodeUint8(data, offset)
	if err != nil {
		return "", 0, err
	}
	offset += SizeOfUint8
	if len(data) < offset+length {
		return "", offset, FatalToShortData
	}
	name := string(data[offset : offset+length])
	offset += length
	return name, offset, nil
}
