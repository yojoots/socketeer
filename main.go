package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type socketSender struct {
	outgoingQueue chan *interface{}
	nextSendTime  time.Time
	connection    net.Conn
}

const (
	defaultProtocol       = "tcp"
	defaultConnectionHost = "127.0.0.1:8137"
	defaultClientPort     = 8138
	protocolVersion       = byte(1)
)

var (
	dialRetryWait = 5 * time.Minute
	redialWait    = 10 * time.Second
	retryLimit    = 3
)

func (socketSender *socketSender) processOutChannel() {
	socketSender.connect()

	for objectToSend := range socketSender.outgoingQueue {
		if socketSender.nextSendTime.IsZero() || socketSender.nextSendTime.Before(time.Now()) {
			socketSender.sendData(objectToSend)
		}
	}
}

func (socketSender *socketSender) sendData(objectToSend interface{}) {
	data, err := json.Marshal(objectToSend)
	if err != nil {
		fmt.Printf("Error while trying to json-marshal data: %v: %v", objectToSend, err)
		return
	}

	// Try to send (at most make retryLimit attempts)
	sendSuccessful := false
	for retry := 0; retry < retryLimit && !sendSuccessful; retry++ {
		if err = socketSender.connect(); err != nil {
			fmt.Printf("Error while connecting to %s: %v, attempt #%d", defaultConnectionHost, err, retry)
			time.Sleep(redialWait)
			continue
		}

		if err = socketSender.writeData(data); err == nil {
			sendSuccessful = true
		} else {
			fmt.Printf("Error while sending message to %s: %v, attempt #%d", defaultConnectionHost, err, retry)

			// reset connection and retry
			socketSender.disconnect()
			socketSender.connection = nil
			time.Sleep(redialWait)
		}
	}

	if !sendSuccessful {
		socketSender.nextSendTime = time.Now().Add(dialRetryWait)
	}
}

func (socketSender *socketSender) connect() error {
	if socketSender.connection == nil {
		var conn net.Conn
		var err error
		conn, err = net.Dial(defaultProtocol, defaultConnectionHost)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		socketSender.connection = conn
		socketSender.nextSendTime = time.Time{}
	}
	return nil
}

func catchConnectPanics() error {
	if r := recover(); r != nil {
		return fmt.Errorf("failed to connect to receiver: %v", r)
	}
	return nil
}

func (socketSender *socketSender) disconnect() {
	if socketSender.connection != nil {
		fmt.Println("Closing connection to receiver", defaultConnectionHost)
		err := socketSender.connection.Close()
		if err != nil {
			fmt.Println("An error occurred while closing connection to receiver", defaultConnectionHost)
		}
	}
}

func (socketSender *socketSender) writeData(data []byte) (err error) {
	defer catchSendPanics()

	writer := bufio.NewWriter(socketSender.connection)
	writer.WriteByte(protocolVersion)
	writer.Flush()

	dataSize := int32(len(data))
	err = binary.Write(writer, binary.LittleEndian, dataSize)
	if err != nil {
		return fmt.Errorf("failed to write data size header: %v", err)
	}

	bytesWritten, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v. Bytes written: %d", err, bytesWritten)
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	return nil
}

func catchSendPanics() error {
	if r := recover(); r != nil {
		return fmt.Errorf("failed to write data: %v", r)
	}
	return nil
}

type Student struct {
	Name string
	Age  int
}

func main() {
	var a Student
	a.Name = "Alice"
	a.Age = 22
	socketSenderInstance := new(socketSender)
	socketSenderInstance.sendData(a)
}
