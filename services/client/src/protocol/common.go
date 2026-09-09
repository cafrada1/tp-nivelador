package protocol

const (
	payloadLengthSize = 4
	messageIDSize     = 4
	winnersLengthSize = 4

	messageTypeSize = 1
)

const (
	messageOpen    byte = 0x00
	messageData    byte = 0x01
	messageClose   byte = 0x02
	messageWinners byte = 0x03
	messageAck     byte = 0x04
)

const (
	betsLengthSize = 2
	nameLengthSize = 1
	documentSize   = 4
	birthdateSize  = 4
	betNumberSize  = 2
)
