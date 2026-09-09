package safe_socket

import "io"

func SendAll(socket io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := socket.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	received := 0
	for received < size {
		n, err := socket.Read(buff[received:])
		if err != nil && (received+n) < size {
			return nil, err
		}
		received += n
	}
	return buff, nil
}
