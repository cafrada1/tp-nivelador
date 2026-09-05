import socket

def recv_all(sock: socket.socket, size: int) -> bytes:
    data: bytearray = bytearray()
    while len(data) < size:
        chunk: bytes = sock.recv(size - len(data))
        if not chunk:
            raise RuntimeError("socket connection broken")
        data.extend(chunk)
    return bytes(data)


def send_all(sock: socket.socket, data: bytes) -> None:
    sent = 0
    while sent < len(data):
        sent += sock.send(data[sent:])
