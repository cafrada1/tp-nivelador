package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type Encoder interface {
	SendBets(bets domain.Bets) error
	SendOpen(agencyId int) error
	SendClose() error
}

type encoder struct {
	sock io.Writer
}

func NewEncoder(sock io.Writer) *encoder {
	return &encoder{sock}
}

func (e *encoder) send(data bytes.Buffer) error {
	payloadContentSize := payloadSize(data)

	var payload bytes.Buffer
	payload.Grow(payloadContentSize)
	err := encodePayloadSize(payloadContentSize, &payload)
	if err != nil {
		return err
	}
	_, err = payload.Write(data.Bytes())

	return safe_socket.SendAll(e.sock, payload.Bytes())
}

func (e *encoder) SendClose() error {
	var data bytes.Buffer
	_, err := data.Write([]byte{messageClose})
	if err != nil {
		return err
	}
	return e.send(data)
}

func (e *encoder) SendOpen(agencyId int) error {
	var data bytes.Buffer
	_, err := data.Write([]byte{messageOpen})
	if err != nil {
		return err
	}
	err = encodeAgencyId(agencyId, &data)
	if err != nil {
		return err
	}
	return e.send(data)
}

func (e *encoder) SendBets(bets domain.Bets) error {
	var data bytes.Buffer

	_, err := data.Write([]byte{messageData})
	if err != nil {
		return err
	}

	err = e.encodeBets(bets, &data)
	if err != nil {
		return err
	}

	return e.send(data)
}

func (e *encoder) encodeBets(bets domain.Bets, data *bytes.Buffer) error {
	data.Grow(betsLengthSize)
	err := encodeLengthBets(len(bets), data)
	if err != nil {
		return err
	}

	for i := range bets {
		err = encodeBet(bets[i], data)
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeAgencyId(agencyId int, data *bytes.Buffer) error {
	return binary.Write(data, binary.BigEndian, uint32(agencyId))
}

func payloadSize(data bytes.Buffer) int {
	return data.Len()
}

func encodePayloadSize(payloadSize int, data *bytes.Buffer) error {
	return binary.Write(data, binary.BigEndian, uint32(payloadSize))
}

func encodeLengthBets(length int, data *bytes.Buffer) error {
	return binary.Write(data, binary.BigEndian, uint16(length))
}

func encodeBet(bet domain.Bet, data *bytes.Buffer) error {
	data.Grow(betSize(bet))

	err := encodeName(bet.FirstName, data)
	if err != nil {
		return err
	}

	err = encodeName(bet.LastName, data)
	if err != nil {
		return err
	}

	err = encodeDocument(bet.Document, data)
	if err != nil {
		return err
	}

	err = encodeBirthdate(bet.Birthdate, data)
	if err != nil {
		return err
	}

	err = encodeBetNumber(bet.Number, data)
	if err != nil {
		return err
	}

	return nil
}

func encodeBetNumber(betNumber int, data *bytes.Buffer) error {
	return binary.Write(data, binary.BigEndian, uint16(betNumber))
}

func encodeBirthdate(birthdate string, data *bytes.Buffer) error {
	var year, month, day int
	_, err := fmt.Sscanf(birthdate, "%d-%d-%d", &year, &month, &day)
	if err != nil {
		return err
	}
	dateInt := year*10000 + month*100 + day

	return binary.Write(data, binary.BigEndian, uint32(dateInt))
}

func encodeDocument(document int, data *bytes.Buffer) error {
	return binary.Write(data, binary.BigEndian, uint32(document))
}

func encodeName(name string, data *bytes.Buffer) error {
	err := binary.Write(data, binary.BigEndian, uint8(len(name)))
	if err != nil {
		return err
	}
	data.WriteString(name)
	return nil
}

func betSize(bet domain.Bet) int {
	return 2*nameLengthSize + len(bet.FirstName) + len(bet.LastName) +
		documentSize + birthdateSize + betNumberSize
}
