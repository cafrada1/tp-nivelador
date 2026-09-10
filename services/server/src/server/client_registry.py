import threading

from server.server_client import ServerClient


class ClientRegistry:
    def __init__(self):
        self._lock: threading.Lock = threading.Lock()
        self._clients: list[ServerClient] = []

    def add(self, client: ServerClient) -> None:
        with self._lock:
            self._remove_close_locked()
            self._clients.append(client)
            client.start()

    def close(self) -> None:
        with self._lock:
            for client in self._clients:
                client.close()
            for client in self._clients:
                client.join()
            self._clients = []

    def _remove_close_locked(self) -> None:
        alive: list[ServerClient] = []
        for client in self._clients:
            if client.is_closed():
                client.join()
            else:
                alive.append(client)
        self._clients = alive
