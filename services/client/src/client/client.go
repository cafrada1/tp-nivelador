package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 10
const CONNECTION_ATTEMPS_DELAY_MS = 300

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
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

func (client *Client) Run() error {
	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "file", client.config.InputFile)
		return err
	}
	defer inputFile.Close()

	input := bufio.NewScanner(inputFile)

	output, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail, "file", client.config.OutputFile)
		return err
	}
	defer output.Close()

	const mainAction = "test-echo-server"
	defer client.conn.Close()

	encoder := protocol.NewEncoder(client.conn)

	agencyId, err := strconv.Atoi(client.config.AgencyId)
	if err != nil {
		logger.Error("parse-agency-id", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	err = encoder.SendOpen(agencyId)
	if err != nil {
		logger.Error("send-open", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	err = client.sendBets(input, encoder)
	if err != nil {
		logger.Error("send-bets", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	err = encoder.SendClose()
	if err != nil {
		logger.Error("send-close", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	decoder := protocol.NewDecoder(client.conn)
	winners, err := decoder.ReceiveWinners()
	if err != nil {
		logger.Error("recv-winners", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	for _, winner := range winners {
		winnerLine := fmt.Sprintf("%s,%s,%d,%s,%d",
			winner.FirstName, winner.LastName, winner.Document, winner.Birthdate, winner.Number)
		if _, err := fmt.Fprintln(output, winnerLine); err != nil {
			logger.Error("write-winner", logger.Fail, "agency-id", agencyId, "winner", winner)
			return err
		}
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return input.Err()
}

func (client *Client) sendBets(input *bufio.Scanner, encoder protocol.Encoder) error {
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

		err = encoder.SendBets(bets)
		if err != nil {
			return err
		}

		i = 0
	}

	if i > 0 {
		err := encoder.SendBets(bets[:i])
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	FirstNameIndex = iota
	LastNameIndex
	DocumentIndex
	BirthdateIndex
	NumberIndex
)

const (
	LineFields = 5
	Separator  = ","
)

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
