package protocol

import (
	"fmt"
	"net"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

type Protocol interface {
	SendOpen(agencyID int) error
	SendBets(bets domain.Bets) error
	SendClose() error
	ReceiveWinners() (domain.Bets, error)
	Close() error
}
type protocol struct {
	conn    net.Conn
	encoder *encoder
	decoder *decoder
	nextID  int
}

func NewProtocol(conn net.Conn) *protocol {
	return &protocol{
		conn:    conn,
		encoder: newEncoder(conn),
		decoder: newDecoder(conn),
		nextID:  0,
	}
}

func (p *protocol) Close() error {
	logger.Info("protocol-close", logger.Success)
	return p.conn.Close()
}

func (p *protocol) ReceiveWinners() (domain.Bets, error) {
	messageID, winners, err := p.receiveWinners()
	if err != nil {
		return nil, err
	}
	return winners, p.SendAck(messageID)
}

func (p *protocol) SendOpen(agencyID int) error {
	messageID, err := p.sendOpen(agencyID)
	if err != nil {
		return err
	}

	return p.receiveAck(messageID)
}

func (p *protocol) SendBets(bets domain.Bets) error {
	messageID, err := p.sendBets(bets)
	if err != nil {
		return err
	}
	return p.receiveAck(messageID)
}

func (p *protocol) SendClose() error {
	messageID, err := p.sendClose()
	if err != nil {
		return err
	}
	return p.receiveAck(messageID)
}
func (p *protocol) nextMessageID() int {
	p.nextID++
	return p.nextID
}

func (p *protocol) sendOpen(agencyID int) (int, error) {
	messageID := p.nextMessageID()
	if err := p.encoder.sendOpen(messageID, agencyID); err != nil {
		return 0, err
	}
	return messageID, nil
}

func (p *protocol) sendBets(bets domain.Bets) (int, error) {
	messageID := p.nextMessageID()
	return messageID, p.encoder.sendBets(messageID, bets)
}

func (p *protocol) sendClose() (int, error) {
	messageID := p.nextMessageID()
	return messageID, p.encoder.sendClose(messageID)
}

func (p *protocol) SendAck(messageID int) error {
	return p.encoder.sendAck(messageID)
}

func (p *protocol) receiveAck(expectedID int) error {
	for {
		messageID, err := p.decoder.receiveAck()
		if err != nil {
			return err
		}
		if messageID == expectedID {
			return nil
		}
		if messageID < expectedID {
			continue
		}
		return fmt.Errorf("protocol error: expected ACK with id %d, got id %d", expectedID, messageID)
	}
}

func (p *protocol) receiveWinners() (int, domain.Bets, error) {
	messageID, winners, err := p.decoder.receiveWinners()
	if err != nil {
		return 0, nil, err
	}
	return messageID, winners, nil
}
