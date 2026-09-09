import socket
import threading

import logger
from server.client_registry import ClientRegistry
from server.lottery_monitor import LotteryMonitor
from server.server_client import ServerClient

class Server(threading.Thread):
    def __init__(
        self, server_host: str, server_port: int, monitor: LotteryMonitor
    ) -> None:
        super().__init__()
        self.server_host: str = server_host
        self.server_port: int = server_port
        self._lottery_monitor: LotteryMonitor = monitor
        self._registry: ClientRegistry = ClientRegistry()
        self._shutdown: threading.Event = threading.Event()
        self._sock: socket.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, True)

    def shutdown(self) -> None:
        if self._shutdown.is_set():
            return
        self._shutdown.set()
        self._sock.shutdown(socket.SHUT_RDWR)
        self._sock.close()

    def run(self):
        action = "accept-connection"

        try:
            self._sock.bind((self.server_host, self.server_port))
            self._sock.listen()
            while not self._shutdown.is_set():
                logger.info(action, logger.LogResult.in_progress)
                client_socket, _ = self._sock.accept()
                logger.info(action, logger.LogResult.success)

                if self._shutdown.is_set():
                    client_socket.close()
                    continue

                self._registry.add(
                    ServerClient(client_socket, self._lottery_monitor)
                )
        except Exception as e:
            logger.error(action, logger.LogResult.fail, "err", e)
        finally:
            self._lottery_monitor.abort()
            self._registry.close()
            self._shutdown.set()
