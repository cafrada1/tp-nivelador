import socket

import logger
from lottery import Bet
from protocol.decoder import Decoder
from protocol.messages import Message
from protocol import MessageType
from protocol.encoder import Encoder

class ProtocolError(Exception):
    pass


class Protocol:
    def __init__(self, sock: socket.socket):
        self._sock: socket.socket = sock
        self._encoder: Encoder = Encoder(sock)
        self._decoder: Decoder = Decoder(sock)
        self._next_id: int = 0
        self._closed: bool = False

    def send_winners(self, winners: list[Bet]) -> None:
        message_id: int = self._send_winners(winners)
        self._recv_ack(message_id)

    def receive_open_message(self) -> Message:
        message_id, message = self._recv_message()
        if message.type != MessageType.OPEN:
            raise ProtocolError(f"Expected OPEN message, got {message.type.name} with id {message_id}")

        self._send_ack(message_id)
        return message

    def receive_data_message(self) -> Message:
        message_id, message = self._recv_message()
        if message.type != MessageType.DATA and message.type != MessageType.CLOSE:
            raise ProtocolError(f"Expected DATA or CLOSE message, got {message.type.name} with id {message_id}")

        self._send_ack(message_id)
        return message

    @property
    def closed(self) -> bool:
        return self._closed

    def close(self):
        if self._closed:
            return
        self._closed = True
        try:
            self._sock.shutdown(socket.SHUT_RDWR)
        except OSError:
            pass
        try:
            self._sock.close()
            logger.info("protocol-close", logger.LogResult.success)
        except OSError as e:
            logger.error("protocol-close", logger.LogResult.fail, "err", e)

    @property
    def agency_id(self) -> int:
        return self._decoder.agency_id

    def _recv_message(self) -> tuple[int, Message]:
        message_id, message = self._decoder.recv_message()
        return message_id, message

    def _send_ack(self, message_id: int) -> None:
        self._encoder.send_ack(message_id)

    def _send_winners(self, winners: list[Bet]) -> int:
        self._next_id += 1
        message_id = self._next_id
        self._encoder.send_winners(winners, message_id)
        return message_id

    def _recv_ack(self, expected_id: int) -> None:
        finish: bool = False
        while not finish:
            message_id, message = self._recv_message()
            if message.type != MessageType.ACK or message_id > expected_id:
                raise ProtocolError(
                    f"Expected ACK with id {expected_id}, got {message.type.name} with id {message_id}"
                )
            finish = message_id == expected_id
