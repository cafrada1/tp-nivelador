package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type encoder struct {
	sock io.Writer
}

func newEncoder(sock io.Writer) *encoder {
	return &encoder{sock}
}

func (e *encoder) send(messageID int, messageType byte, data *bytes.Buffer) error {
	var payload bytes.Buffer
	payload.Grow(messageIDSize + messageTypeSize + data.Len())
	if err := binary.Write(&payload, binary.BigEndian, uint32(messageID)); err != nil {
		return err
	}
	if err := payload.WriteByte(messageType); err != nil {
		return err
	}
	if _, err := payload.Write(data.Bytes()); err != nil {
		return err
	}

	var frame bytes.Buffer
	frame.Grow(payloadLengthSize + payload.Len())
	if err := binary.Write(&frame, binary.BigEndian, uint32(payload.Len())); err != nil {
		return err
	}
	if _, err := frame.Write(payload.Bytes()); err != nil {
		return err
	}

	return safe_socket.SendAll(e.sock, frame.Bytes())
}

func (e *encoder) sendAck(messageID int) error {
	var data bytes.Buffer
	return e.send(messageID, messageAck, &data)
}

func (e *encoder) sendClose(messageID int) error {
	var data bytes.Buffer
	return e.send(messageID, messageClose, &data)
}

func (e *encoder) sendOpen(messageID int, agencyID int) error {
	var data bytes.Buffer
	if err := encodeAgencyID(agencyID, &data); err != nil {
		return err
	}
	return e.send(messageID, messageOpen, &data)
}

func (e *encoder) sendBets(messageID int, bets domain.Bets) error {
	var data bytes.Buffer

	if err := e.encodeBets(bets, &data); err != nil {
		return err
	}

	return e.send(messageID, messageData, &data)
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

func encodeAgencyID(agencyID int, data *bytes.Buffer) error {
	return binary.Write(data, binary.BigEndian, uint32(agencyID))
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
