# Socketeer

Proof-of-concept socket-based communication examples, implemented in different languages.

## Go

The `main.go` file contains code for a minimal transmitter application, which will dump an example JSON blob over the wire to any connected listeners. A straightforward `go build` or `go install` will produce the `socketeer` binary which can be run at will.

## Python

Both `sender.py` and `receiver.py` files are included. Nothing too fancy here. For best results, use python3.