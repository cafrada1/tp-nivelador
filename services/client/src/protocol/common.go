package protocol

const (
	payloadLengthSize = 4
	winnersLengthSize = 4
)

const (
	messageOpen  byte = 0x00
	messageData  byte = 0x01
	messageClose byte = 0x02
)

const (
	betsLengthSize = 2
	nameLengthSize = 1
	documentSize   = 4
	birthdateSize  = 4
	betNumberSize  = 2
)
