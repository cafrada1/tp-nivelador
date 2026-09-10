import socket

def recv_all(sock: socket.socket, size: int) -> bytes:
    """Lee exactamente size bytes, tolerando lecturas parciales del socket."""
    data: bytearray = bytearray()
    while len(data) < size:
        chunk: bytes = sock.recv(size - len(data))
        if not chunk:
            raise RuntimeError("socket connection broken")
        data.extend(chunk)
    return bytes(data)


def send_all(sock: socket.socket, data: bytes) -> None:
    """Envia todo el buffer, tolerando escrituras parciales del socket."""
    sent = 0
    while sent < len(data):
        sent += sock.send(data[sent:])
