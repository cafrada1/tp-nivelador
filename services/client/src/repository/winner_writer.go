package repository

import (
	"fmt"
	"os"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
)

const (
	lineFormat = "%s,%s,%d,%s,%d\n"
)

type WinnerWriter interface {
	WriteWinners(domain.Bets) error
	Close() error
}

type winnerWriter struct {
	file *os.File
}

func NewWinnerWriter(path string) (*winnerWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &winnerWriter{file: file}, nil
}

func (w *winnerWriter) Close() error {
	return w.file.Close()
}

func (w *winnerWriter) WriteWinners(winners domain.Bets) error {
	for _, winner := range winners {
		winnerLine := fmt.Sprintf(lineFormat,
			winner.FirstName, winner.LastName, winner.Document, winner.Birthdate, winner.Number)
		n, err := w.file.WriteString(winnerLine)
		if err != nil {
			return err
		}
		for n < len(winnerLine) {
			// No deberia escribirse parcialmente sin un error pero por las dudas se escribe lo faltante.
			m, err := w.file.WriteString(winnerLine[n:])
			if err != nil {
				return err
			}
			n += m
		}
	}
	return nil
}
