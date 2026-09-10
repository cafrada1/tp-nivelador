package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

type Sender interface {
	SendAck(messageID int) error
	SendOpen(messageID, agencyID int) error
	SendClose(messageID int) error
	SendBets(messageID int, bets domain.Bets) error
}
type sender struct {
	sock io.Writer
}

func NewSender(sock io.Writer) *sender {
	return &sender{sock}
}

func (e *sender) SendAck(messageID int) error {
	return e.send(messageID, messageAck, nil)
}

func (e *sender) SendClose(messageID int) error {
	return e.send(messageID, messageClose, nil)
}

func (e *sender) SendOpen(messageID int, agencyID int) error {
	var data bytes.Buffer
	if err := encodeAgencyID(agencyID, &data); err != nil {
		return err
	}

	encodeFunc := func(data *bytes.Buffer) error {
		return encodeAgencyID(agencyID, data)
	}
	return e.send(messageID, messageOpen, encodeFunc)
}

func (e *sender) SendBets(messageID int, bets domain.Bets) error {
	var data bytes.Buffer

	if err := e.encodeBets(bets, &data); err != nil {
		return err
	}

	encodeFunc := func(data *bytes.Buffer) error {
		return e.encodeBets(bets, data)
	}

	return e.send(messageID, messageData, encodeFunc)
}

func (e *sender) send(messageID int, messageType byte, encodeBody func(*bytes.Buffer) error) error {
	var frame bytes.Buffer
	frame.Grow(payloadLengthSize + messageIDSize + messageTypeSize)
	frame.Write([]byte{0, 0, 0, 0}) // placeholder para payloadLength

	if err := EncodeUint32(&frame, messageID); err != nil {
		return err
	}
	if err := frame.WriteByte(messageType); err != nil {
		return err
	}
	if encodeBody != nil {
		if err := encodeBody(&frame); err != nil {
			return err
		}
	}

	length := frame.Len() - payloadLengthSize
	binary.BigEndian.PutUint32(frame.Bytes(), uint32(length))
	return safe_socket.SendAll(e.sock, frame.Bytes())
}

func (e *sender) encodeBets(bets domain.Bets, data *bytes.Buffer) error {
	data.Grow(betsLengthSize)

	if err := encodeLengthBets(len(bets), data); err != nil {
		return err
	}

	for i := range bets {
		if err := encodeBet(bets[i], data); err != nil {
			return err
		}
	}
	return nil
}

func encodeAgencyID(agencyID int, data *bytes.Buffer) error {
	if err := EncodeUint32(data, agencyID); err != nil {
		return fmt.Errorf("cannot encode agency ID %d: %w", agencyID, err)
	}
	return nil
}

func encodeLengthBets(length int, data *bytes.Buffer) error {
	if err := EncodeUint16(data, length); err != nil {
		return fmt.Errorf("cannot encoding length bets %d: %w", length, err)
	}
	return nil
}

func encodeBet(bet domain.Bet, data *bytes.Buffer) error {
	data.Grow(betSize(bet))

	if err := encodeName(bet.FirstName, data); err != nil {
		return err
	}

	if err := encodeName(bet.LastName, data); err != nil {
		return err
	}

	if err := encodeDocument(bet.Document, data); err != nil {
		return err
	}

	if err := encodeBirthdate(bet.Birthdate, data); err != nil {
		return err
	}

	if err := encodeBetNumber(bet.Number, data); err != nil {
		return err
	}

	return nil
}

func encodeBetNumber(betNumber int, data *bytes.Buffer) error {
	err := EncodeUint32(data, betNumber)
	if err != nil {
		return fmt.Errorf("cannot encode bet number %d: %w", betNumber, err)
	}
	return nil
}

func encodeBirthdate(birthdate string, data *bytes.Buffer) error {
	var year, month, day int
	_, err := fmt.Sscanf(birthdate, "%d-%d-%d", &year, &month, &day)
	if err != nil {
		return err
	}
	dateInt := year*10000 + month*100 + day

	err = EncodeUint32(data, dateInt)
	if err != nil {
		return fmt.Errorf("cannot encode birthdate %s: %w", birthdate, err)
	}
	return nil
}

func encodeDocument(document int, data *bytes.Buffer) error {
	err := EncodeUint32(data, document)
	if err != nil {
		return fmt.Errorf("cannot encode document %d: %w", document, err)
	}
	return nil
}

func encodeName(name string, data *bytes.Buffer) error {
	err := EncodeUint8(data, len(name))
	if err != nil {
		return fmt.Errorf("cannot write name %s: %w", name, err)
	}
	data.WriteString(name)
	return nil
}

func betSize(bet domain.Bet) int {
	return 2*nameLengthSize + len(bet.FirstName) + len(bet.LastName) +
		documentSize + birthdateSize + betNumberSize
}
