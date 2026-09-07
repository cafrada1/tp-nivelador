from socket import socket

from protocol import common
from protocol.messages import Message, OpenMessage, DataMessage, CloseMessage, AckMessage, MessageType
from safe_socket import safe_socket
from lottery import Bet

class MalformedMessageError(Exception):
    pass


class Decoder:
    def __init__(self, sock: socket):
        self._agency_id: int | None = None
        self._sock: socket = sock

    @property
    def agency_id(self) -> int:
        if self._agency_id is None:
            raise ValueError("Agency ID not set")
        return self._agency_id

    @agency_id.setter
    def agency_id(self, value: int) -> None:
        self._agency_id = value

    def recv_message(self) -> tuple[int, Message]:
        message_id, message, data = self._recv_payload()

        match message:
            case MessageType.OPEN.value:
                self._agency_id = _decode_agency_id(data)
                return message_id, OpenMessage(self.agency_id)
            case MessageType.DATA.value:
                return message_id, DataMessage(_decode_bets(self.agency_id, data))
            case MessageType.CLOSE.value:
                return message_id, CloseMessage()
            case MessageType.ACK.value:
                return message_id, AckMessage()
            case _:
                raise MalformedMessageError

    def _recv_payload(self) -> tuple[int, int, bytes]:
        payload_size: bytes = safe_socket.recv_all(self._sock, common.PAYLOAD_LENGTH_SIZE)
        payload_size_int: int = int.from_bytes(payload_size, byteorder=common.ENDIAN)
        if payload_size_int < common.MESSAGE_TYPE_SIZE + common.MESSAGE_ID_SIZE:
            raise MalformedMessageError

        payload: bytes = safe_socket.recv_all(self._sock, payload_size_int)

        message_id: int = int.from_bytes(payload[:common.MESSAGE_ID_SIZE], byteorder=common.ENDIAN)
        message: int = payload[common.MESSAGE_ID_SIZE]
        data: bytes = payload[common.MESSAGE_ID_SIZE + common.MESSAGE_TYPE_SIZE:]
        return message_id, message, data

def _decode_agency_id(data: bytes) -> int:
    agency_id: int = int.from_bytes(data[:common.AGENCY_ID_SIZE], byteorder=common.ENDIAN)
    return agency_id

def _decode_bets(agency_id: int, data: bytes) -> list[Bet]:
    bets: list[Bet] = []
    size: int = int.from_bytes(data[:common.BETS_LENGTH_SIZE], byteorder=common.ENDIAN)

    offset: int = common.BETS_LENGTH_SIZE
    for _ in range(size):
        bet, offset = _decode_bet(agency_id, data, offset)
        bets.append(bet)
    return bets

def _decode_bet(agency_id: int, data: bytes, offset: int) -> tuple[Bet, int]:
    first_name, offset = _decode_name(data, offset)

    last_name, offset = _decode_name(data, offset)

    document, offset = _decode_document(data, offset)

    date_of_birth, offset = _decode_date_of_birth(data, offset)

    number, offset = _decode_bet_number(data, offset)

    bet: Bet = Bet(first_name=first_name, last_name=last_name,
               agency_id=agency_id, document=document, birthdate=date_of_birth, number=number)

    return bet, offset


def _decode_bet_number(data: bytes, offset: int) -> tuple[int, int]:
    bet: int = int.from_bytes(data[offset:offset + common.BET_NUMBER_SIZE], byteorder=common.ENDIAN)
    offset += common.BET_NUMBER_SIZE
    return bet, offset

def _decode_document(data: bytes, offset: int) -> tuple[int, int]:
    document: int = int.from_bytes(data[offset:offset + common.DOCUMENT_SIZE], byteorder=common.ENDIAN)
    offset += common.DOCUMENT_SIZE
    return document, offset

def _decode_date_of_birth(data: bytes, offset: int) -> tuple[str, int]:
    date_of_birth_bytes: bytes = data[offset:offset + common.BIRTH_DATE_SIZE]
    date_of_birth_int: int = int.from_bytes(date_of_birth_bytes, byteorder=common.ENDIAN)
    offset += common.BIRTH_DATE_SIZE
    return _birth_date_from_int(date_of_birth_int), offset


def _birth_date_from_int(date_int: int) -> str:
    year: int = date_int // 10000
    month: int = (date_int // 100) % 100
    day: int = date_int % 100
    return f"{year:04d}-{month:02d}-{day:02d}"


def _decode_name(data: bytes, offset: int = 0) -> tuple[str, int]:
    name_length: int = int.from_bytes(data[offset:offset + common.NAME_LENGTH_SIZE], byteorder=common.ENDIAN)
    offset += common.NAME_LENGTH_SIZE
    name: str = data[offset:offset + name_length].decode("utf-8")
    offset += name_length
    return name, offset
