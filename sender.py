import json
import socket
import struct
import sys

HOST = '127.0.0.1'    # The remote host
PORT = 8137           # The same port as used by the server
protocol_version = struct.pack(">i", 1)
s = None
for res in socket.getaddrinfo(HOST, PORT, socket.AF_UNSPEC, socket.SOCK_STREAM):
    af, socktype, proto, canonname, sa = res
    try:
        s = socket.socket(af, socktype, proto)
    except socket.error as e:
        s = None
        continue
    try:
        s.connect(sa)
    except socket.error as e:
        s.close()
        s = None
        continue
    break
if s is None:
    print('could not open socket')
    sys.exit(1)

object_to_send = {"Age": 24, "Name": "Bob"}
bytes_to_send = str.encode(json.dumps(object_to_send))
msg_len = struct.pack("<i", len(bytes_to_send))

s.send(protocol_version+msg_len+bytes_to_send)
s.close()