package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 10
const CONNECTION_ATTEMPS_DELAY_MS = 300

const (
	FirstNameIndex = iota
	LastNameIndex
	DocumentIndex
	BirthdateIndex
	NumberIndex

	LineFields = 5
	Separator  = ","
)

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   int
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	config     ClientConfig
	protocol   protocol.Protocol
	inputFile  *os.File
	outputFile *os.File
	isClosed   atomic.Bool
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail, "host", config.ServerHost, "port", config.ServerPort, "err", err)
		return nil, err
	}

	inputFile, err := os.Open(config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "file", config.InputFile, "err", err)
		conn.Close()
		return nil, err
	}

	output, err := os.Create(config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail, "file", config.OutputFile, "err", err)
		conn.Close()
		inputFile.Close()
		return nil, err
	}

	client := &Client{
		config:     config,
		protocol:   protocol.NewProtocol(conn),
		inputFile:  inputFile,
		outputFile: output,
	}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
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
	client.inputFile.Close()
	client.outputFile.Close()
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
	const mainAction = "test-echo-server"

	logger.Info("send-open", logger.InProgress, "agency-id", client.config.AgencyId)
	if err := client.protocol.SendOpen(client.config.AgencyId); err != nil {
		logger.Error("send-open", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}
	logger.Info("send-open", logger.Success, "agency-id", client.config.AgencyId)

	logger.Info(mainAction, logger.InProgress, "agency-id", client.config.AgencyId)
	if err := client.sendBets(); err != nil {
		logger.Error("send-bets", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}
	logger.Info("send-bets", logger.Success, "agency-id", client.config.AgencyId)

	logger.Info("send-close", logger.InProgress, "agency-id", client.config.AgencyId)
	if err := client.protocol.SendClose(); err != nil {
		logger.Error("send-close", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}
	logger.Info("send-close", logger.Success, "agency-id", client.config.AgencyId)

	logger.Info("process-winners", logger.InProgress, "agency-id", client.config.AgencyId)
	if err := client.processWinners(); err != nil {
		logger.Error("process-winners", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
		return err
	}
	logger.Info("process-winners", logger.Success, "agency-id", client.config.AgencyId)

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

func (client *Client) saveWinners(winners domain.Bets, agencyId int) error {
	for _, winner := range winners {
		winnerLine := fmt.Sprintf("%s,%s,%d,%s,%d",
			winner.FirstName, winner.LastName, winner.Document, winner.Birthdate, winner.Number)
		if _, err := fmt.Fprintln(client.outputFile, winnerLine); err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) processWinners() error {
	winners, err := client.protocol.ReceiveWinners()
	if err != nil {
		return err
	}

	if err = client.saveWinners(winners, client.config.AgencyId); err != nil {
		return err
	}
	return nil
}

func (client *Client) sendBets() error {
	input := bufio.NewScanner(client.inputFile)
	bets := make(domain.Bets, client.config.BatchSize)
	i := 0
	for input.Scan() {
		line := input.Text()
		if line == "" {
			continue
		}

		bet, err := parseLine(line)
		if err != nil {
			return err
		}

		bets[i] = bet
		i++

		if i < client.config.BatchSize {
			continue
		}

		if err = client.protocol.SendBets(bets); err != nil {
			return err
		}
		i = 0
	}
	if i > 0 {
		return client.protocol.SendBets(bets[:i])
	}

	return input.Err()
}

func parseLine(line string) (domain.Bet, error) {
	values := strings.Split(line, Separator)
	if len(values) != LineFields {
		return domain.Bet{}, fmt.Errorf("invalid bet line: %s", line)
	}

	document, err := strconv.Atoi(values[DocumentIndex])
	if err != nil {
		return domain.Bet{}, err
	}

	number, err := strconv.Atoi(values[NumberIndex])
	if err != nil {
		return domain.Bet{}, err
	}

	return domain.Bet{
		FirstName: values[FirstNameIndex],
		LastName:  values[LastNameIndex],
		Document:  document,
		Birthdate: values[BirthdateIndex],
		Number:    number,
	}, nil
}
