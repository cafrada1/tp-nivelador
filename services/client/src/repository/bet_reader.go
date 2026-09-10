package repository

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
)

const (
	firstNameIndex = iota
	lastNameIndex
	documentIndex
	birthdateIndex
	numberIndex

	lineFields = 5
	separator  = ","
)

const (
	// Cada cuántos bytes leídos se libera la page cache del archivo.
	dropCacheChunkBytes = 1 << 17
	// POSIX_FADV_DONTNEED: las páginas limpias del rango se evictionan.
	posixFadvDontNeed = 4
)

type BetReader interface {
	ReadUntil(n int) (domain.Bets, error)
	End() bool
	Close() error
}

type betRepository struct {
	file   *os.File
	reader *bufio.Scanner
	end    bool

	readBytes     int
	dropCacheUpTo int
}

func NewBetReader(path string) (*betRepository, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := betRepository{
		file:   file,
		reader: bufio.NewScanner(file),
		end:    false,
	}
	return &reader, nil
}

func (r *betRepository) Close() error {
	return r.file.Close()
}

func (r *betRepository) read() (domain.Bet, error) {
	if !r.reader.Scan() {
		// Al agotar el archivo marca end
		// El EOF no se propaga como error para distinguirlo de fallos reales de lectura.
		r.end = true
		if err := r.reader.Err(); err != nil {
			return domain.Bet{}, err
		}
		return domain.Bet{}, io.EOF
	}
	return r.readLine()
}

// ReadUntil Lee hasta n apuestas.
// Si el archivo termina antes, devuelve el lote parcial junto con end=true en lugar de un error.
func (r *betRepository) ReadUntil(n int) (domain.Bets, error) {
	bets := make(domain.Bets, n)

	var i int
	for i = 0; !r.end && i < n; i++ {
		bet, err := r.read()
		if errors.Is(err, io.EOF) {
			return bets[:i], nil
		}
		if err != nil {
			return nil, err
		}

		bets[i] = bet
	}
	return bets[:i], nil
}

func (r *betRepository) End() bool {
	return r.end
}

func (r *betRepository) readLine() (domain.Bet, error) {
	line := r.reader.Bytes()
	r.trackRead(len(line) + 1)
	return parseLine(string(line))
}

// trackRead acumula los bytes consumidos y, cada dropCacheChunkBytes,
// libera la page cache ya leída para que el pico de memoria del contenedor
// no crezca con el tamaño del archivo de entrada.
func (r *betRepository) trackRead(n int) {
	r.readBytes += n
	if r.readBytes < r.dropCacheUpTo {
		return
	}
	r.dropCacheUpTo = r.readBytes + dropCacheChunkBytes
	syscall.Syscall6(syscall.SYS_FADVISE64, r.file.Fd(), 0, uintptr(r.readBytes), posixFadvDontNeed, 0, 0)
}

func parseLine(line string) (domain.Bet, error) {
	values := strings.Split(line, separator)
	if len(values) != lineFields {
		return domain.Bet{}, fmt.Errorf("invalid bet line: %s", line)
	}

	document, err := strconv.Atoi(values[documentIndex])
	if err != nil {
		return domain.Bet{}, err
	}

	number, err := strconv.Atoi(values[numberIndex])
	if err != nil {
		return domain.Bet{}, err
	}

	return domain.Bet{
		FirstName: values[firstNameIndex],
		LastName:  values[lastNameIndex],
		Document:  document,
		Birthdate: values[birthdateIndex],
		Number:    number,
	}, nil
}
