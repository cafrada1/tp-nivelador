package client

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/repository"
)

const ConnectionAttemptsMax = 10
const ConnectionAttemptsDelayMs = 300

const (
	agencyIdTag = "agency-id"
	errorTag    = "error"
)

type Client struct {
	config        Config
	protocol      protocol.BetProtocol
	betsReader    repository.BetReader
	winnersWriter repository.WinnerWriter
	isClosed      atomic.Bool
}

func NewClient(config Config) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail, "host", config.ServerHost, "port", config.ServerPort, errorTag, err)
		return nil, err
	}

	reader, err := repository.NewBetReader(config.InputFile)
	if err != nil {
		logger.Error("create-bet-reader", logger.Fail, "file", config.InputFile, errorTag, err)
		conn.Close()
		return nil, err
	}

	writer, err := repository.NewWinnerWriter(config.OutputFile)
	if err != nil {
		logger.Error("create-winner-writer", logger.Fail, "file", config.OutputFile, errorTag, err)
		conn.Close()
		reader.Close()
		return nil, err
	}

	client := &Client{
		config:        config,
		protocol:      protocol.NewProtocol(conn),
		betsReader:    reader,
		winnersWriter: writer,
	}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range ConnectionAttemptsMax {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(ConnectionAttemptsDelayMs * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Close() {
	if client.isClosed.Swap(true) {
		return
	}
	client.protocol.Close()
	client.betsReader.Close()
	client.winnersWriter.Close()
}

func (client *Client) Run() error {
	defer client.Close()

	err := client.processBets()
	if !client.isClosed.Load() && err != nil {
		return err
	}
	return nil
}

func (client *Client) processBets() error {
	const (
		clientTag         = "client-bets"
		processWinnersTag = "process-winners"
	)
	logger.Info(clientTag, logger.InProgress, agencyIdTag, client.config.AgencyId)
	err := client.sendAllBets()
	if err != nil {
		return err
	}

	logger.Info(processWinnersTag, logger.InProgress, agencyIdTag, client.config.AgencyId)
	if err = client.processWinners(); err != nil {
		logger.Error(processWinnersTag, logger.Fail, agencyIdTag, client.config.AgencyId, errorTag, err)
		return err
	}
	logger.Info(processWinnersTag, logger.Success, agencyIdTag, client.config.AgencyId)

	logger.Info(clientTag, logger.Success, agencyIdTag, client.config.AgencyId)
	return nil
}

func (client *Client) sendAllBets() error {
	const (
		actionSendOpen  = "send-open"
		actionSendClose = "send-close"
		actionSendBets  = "send-bets"
	)

	logger.Info(actionSendOpen, logger.InProgress, agencyIdTag, client.config.AgencyId)
	if err := client.protocol.SendOpen(client.config.AgencyId); err != nil {
		logger.Error(actionSendOpen, logger.Fail, agencyIdTag, client.config.AgencyId, errorTag, err)
		return err
	}
	logger.Info(actionSendOpen, logger.Success, agencyIdTag, client.config.AgencyId)

	logger.Info(actionSendBets, logger.InProgress, agencyIdTag, client.config.AgencyId)
	if err := client.sendBets(); err != nil {
		logger.Error(actionSendBets, logger.Fail, agencyIdTag, client.config.AgencyId, errorTag, err)
		return err
	}
	logger.Info(actionSendBets, logger.Success, agencyIdTag, client.config.AgencyId)

	logger.Info(actionSendClose, logger.InProgress, agencyIdTag, client.config.AgencyId)
	if err := client.protocol.SendClose(); err != nil {
		logger.Error(actionSendClose, logger.Fail, agencyIdTag, client.config.AgencyId, errorTag, err)
		return err
	}
	logger.Info(actionSendClose, logger.Success, agencyIdTag, client.config.AgencyId)

	return nil
}

func (client *Client) processWinners() error {
	winners, err := client.protocol.ReceiveWinners()
	if err != nil {
		return err
	}

	if err = client.winnersWriter.WriteWinners(winners); err != nil {
		return err
	}
	return nil
}

func (client *Client) sendBets() error {
	for !client.betsReader.End() {
		bets, err := client.betsReader.ReadUntil(client.config.BatchSize)
		if err != nil {
			return err
		}

		if err = client.protocol.SendBets(bets); err != nil {
			return err
		}
	}
	return nil
}
