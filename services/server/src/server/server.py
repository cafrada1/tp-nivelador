import socket

import logger
from protocol import Decoder, MessageType
from protocol.encoder import Encoder
from lottery import Lottery, Bet
from server.lottery_monitor import LotteryMonitor

_ECHO_SERVER_MESSAGE_SIZE = 1024


class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = LotteryMonitor(Lottery(storage_path))

    def _handle_client(self, client_socket):
        decoder: Decoder = Decoder(client_socket)
        winners: list[Bet] = self._recv_bets(decoder)
        encoder: Encoder = Encoder(client_socket)
        encoder.send_message(winners)

    def _recv_bets(self, decoder: Decoder) -> list[Bet]:
        action = "handle-client"
        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)

            message = decoder.recv_message()
            if message.type != MessageType.OPEN:
                logger.error(
                    action, logger.LogResult.fail, "bad-first-open", message.type
                )
                raise Exception("First message must be OPEN")

            message = decoder.recv_message()
            while message.type != MessageType.CLOSE:
                if message.type != MessageType.DATA:
                    logger.error(
                        action, logger.LogResult.fail, "bad-bets-data", message.type
                    )
                    raise Exception("Expected BET message")

                self.lottery.store_bets(message.bets())
                message = decoder.recv_message()
                message_amount += 1

        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            raise e

        return self.lottery.winners(decoder.agency_id)

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
