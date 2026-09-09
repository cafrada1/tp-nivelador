import socket
import threading

import logger
from lottery import Bet
from protocol.receiver import Receiver
from protocol.messages import Message
from protocol import MessageType
from protocol.sender import Sender

class ProtocolError(Exception):
    pass


class Protocol:
    def __init__(self, sock: socket.socket):
        self._sock: socket.socket = sock
        self._sender: Sender = Sender(sock)
        self._receiver: Receiver = Receiver(sock)
        self._next_id: int = 0
        self._closed: threading.Event = threading.Event()

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
        return self._closed.is_set()

    def close(self):
        if self._closed.is_set():
            return
        self._closed.set()
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
        return self._receiver.agency_id

    def _recv_message(self) -> tuple[int, Message]:
        message_id, message = self._receiver.recv_message()
        return message_id, message

    def _send_ack(self, message_id: int) -> None:
        self._sender.send_ack(message_id)

    def _send_winners(self, winners: list[Bet]) -> int:
        self._next_id += 1
        message_id = self._next_id
        self._sender.send_winners(winners, message_id)
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
