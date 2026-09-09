package protocol

import (
	"fmt"
	"net"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

type BetProtocol interface {
	SendOpen(agencyID int) error
	SendBets(bets domain.Bets) error
	SendClose() error
	ReceiveWinners() (domain.Bets, error)
	Close() error
}
type protocol struct {
	conn     net.Conn
	sender   Sender
	receiver Receiver
	nextID   int
}

func NewProtocol(conn net.Conn) *protocol {
	return &protocol{
		conn:     conn,
		sender:   NewSender(conn),
		receiver: NewReceiver(conn),
		nextID:   0,
	}
}

func (p *protocol) Close() error {
	logger.Info("protocol-close", logger.Success)
	return p.conn.Close()
}

func (p *protocol) ReceiveWinners() (domain.Bets, error) {
	winnersMsg, err := p.receiver.ReceiveWinners()
	if err != nil {
		return nil, err
	}
	return winnersMsg.Winners, p.sender.SendAck(winnersMsg.Id)
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
	if err := p.sender.SendOpen(messageID, agencyID); err != nil {
		return 0, err
	}
	return messageID, nil
}

func (p *protocol) sendBets(bets domain.Bets) (int, error) {
	messageID := p.nextMessageID()
	return messageID, p.sender.SendBets(messageID, bets)
}

func (p *protocol) sendClose() (int, error) {
	messageID := p.nextMessageID()
	return messageID, p.sender.SendClose(messageID)
}

func (p *protocol) receiveAck(expectedID int) error {
	for {
		messageID, err := p.receiver.ReceiveAck()
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
