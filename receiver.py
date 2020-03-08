import errno
import json
import socket
import struct
import sys

HOST = '127.0.0.1'      # Default localhost
PORT = 8137             # Arbitrary non-privileged port
s = None

def start_listening():
    print("starting to listen")
    for res in socket.getaddrinfo(HOST, PORT, socket.AF_UNSPEC,
                                  socket.SOCK_STREAM, 0, socket.AI_PASSIVE):
        af, socktype, proto, canonname, sa = res
        try:
            s = socket.socket(af, socktype, proto)
            s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        except socket.error as msg:
            s = None
            continue
        try:
            s.bind(sa)
            s.listen(1)
        except socket.error as msg:
            s.close()
            s = None
            print(msg)
            continue
        break
    if s is None:
        print('could not open socket')
        sys.exit(1)
    while True:
        conn, addr = s.accept()
        print('connected by', addr)
        while True:
            try: 
                protocol_version = conn.recv(1)
                if not protocol_version:
                    break
                #conn.sendall(protocol_version)
                if protocol_version[0] == 1:
                    next_message_size_bytes = conn.recv(4)
                    next_message_size = struct.unpack("<i", next_message_size_bytes)[0]
                    if next_message_size is not None:
                        message_data = conn.recv(next_message_size)
                        remaining_bytes = next_message_size - len(message_data)
                        while remaining_bytes > 0:
                            message_data += conn.recv(remaining_bytes)
                            remaining_bytes = next_message_size - len(message_data)
                        if not message_data: break
                        print(message_data)
            except socket.error as e:
                if e.errno != errno.ECONNRESET:
                    raise # Not error we are looking for
                print(f"Error: {e}")
    conn.close()
    s.close()
    print("connection closed")

start_listening()
