from protocol import common
from safe_socket import safe_socket
from lottery import Bet

class Encoder:
    def __init__(self, sock):
        self._sock = sock

    def send_message(self, winners: list[Bet]) -> None:
        winners_count: int = len(winners)
        payload: bytes = winners_count.to_bytes(common.WINNER_LENGTH_SIZE, byteorder=common.ENDIAN)

        for winner in winners:
            payload += _encode_winner(winner)

        payload_size: int = len(payload)
        payload_size_bytes: bytes = payload_size.to_bytes(common.PAYLOAD_LENGTH_SIZE, byteorder=common.ENDIAN)

        message: bytes = payload_size_bytes + payload
        safe_socket.send_all(self._sock, message)

def _encode_winner(winner: Bet) -> bytes:
    data: bytes = b""

    data += _encode_name(winner.first_name)
    data += _encode_name(winner.last_name)
    data += winner.document.to_bytes(common.DOCUMENT_SIZE, byteorder=common.ENDIAN)
    data += _encode_birthdate(winner.birthdate)
    data += winner.number.to_bytes(common.BET_NUMBER_SIZE, byteorder=common.ENDIAN)

    return data


def _encode_name(name: str) -> bytes:
    name_bytes: bytes = name.encode("utf-8")
    name_length: int = len(name_bytes)
    name_length_bytes: bytes = name_length.to_bytes(common.NAME_LENGTH_SIZE, byteorder=common.ENDIAN)
    return name_length_bytes + name_bytes

def _encode_birthdate(birthdate: str) -> bytes:
    year, month, day = map(int, birthdate.split("-"))
    birthdate_int: int = year * 10000 + month * 100 + day
    return birthdate_int.to_bytes(common.BIRTH_DATE_SIZE, byteorder=common.ENDIAN)