import socket
import threading

import logger
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
        self._clients: list[ServerClient] = []

    def run(self):
        action = "accept-connection"
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
                server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                server_socket.bind((self.server_host, self.server_port))
                server_socket.listen()
                while True:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                    logger.info(action, logger.LogResult.success)

                    client = ServerClient(client_socket, self._lottery_monitor)
                    self._clients.append(client)
                    client.start()
        except Exception as e:
            logger.error("server-run", logger.LogResult.fail, "err", e)
        finally:
            self._close_clients()

    def _close_clients(self):
        for client in self._clients:
            client.close()
            client.join()
