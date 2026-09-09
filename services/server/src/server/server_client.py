import socket
import threading

import logger
from protocol import Protocol, MessageType
from server.lottery_monitor import LotteryMonitor


class ServerClient(threading.Thread):
    def __init__(self, client_socket: socket.socket, lottery_monitor: LotteryMonitor):
        super().__init__(daemon=True)
        self._protocol: Protocol = Protocol(client_socket)
        self._lottery_monitor: LotteryMonitor = lottery_monitor
        self._closed: threading.Event = threading.Event()

    def is_closed(self) -> bool:
        return self._closed.is_set()

    def close(self) -> None:
        if self._closed.is_set():
            return
        self._closed.set()
        self._protocol.close()

    def run(self) -> None:
        action = "handle-client"
        try:
            logger.info(action, logger.LogResult.in_progress)
            self._process_bets()

            winners = self._lottery_monitor.winners(self._protocol.agency_id)
            self._protocol.send_winners(winners)

            logger.info(
                action,
                logger.LogResult.success,
                "agency-id",
                self._protocol.agency_id,
                "winners-amount",
                len(winners),
            )
        except Exception as e:
            if not self.is_closed():
                logger.error(action, logger.LogResult.fail, "err", e)
        finally:
            self.close()

    def _process_bets(self) -> None:
        batches_amount = 0

        self._protocol.receive_open_message()

        logger.info("recv-open", logger.LogResult.success, "agency-id", self._protocol.agency_id)

        message = self._protocol.receive_data_message()
        while message.type != MessageType.CLOSE:
            self._lottery_monitor.store_bets(message.bets())
            batches_amount += 1

            message = self._protocol.receive_data_message()

        logger.info(
            "recv-bets",
            logger.LogResult.success,
            "agency-id",
            self._protocol.agency_id,
            "batches-amount",
            batches_amount,
        )
