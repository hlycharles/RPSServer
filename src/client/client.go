package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"../rps"
)

const port = ":8888"

type clientState int

const (
	clientWait clientState = iota
	clientConnected
	clientInGame
)

type client struct {
	conn         net.Conn
	currentScore int
	state        clientState
	id           int
	closeChan    chan bool
}

// Connect to server.
func main() {
	conn, err := net.Dial("tcp", port)
	if err != nil {
		fmt.Println("Unable to connect to server")
		return
	}
	c := client{
		conn:         conn,
		currentScore: 0,
		state:        clientWait,
		closeChan:    make(chan bool),
	}
	c.handleConnection()
	<-c.closeChan
}

// Handle connection to server.
func (c *client) handleConnection() {
	go c.handleServerMessage()
	go c.handleClientInput()
}

// Handle client input.
func (c *client) handleClientInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Unable to read input")
			continue
		}
		var m rps.Message
		switch c.state {
		case clientConnected:
			m = rps.Message{
				MsgType: rps.MsgStart,
			}
		case clientInGame:
			// remove newline character
			text = text[:len(text)-1]
			m = rps.Message{
				MsgType:    rps.MsgMove,
				MsgContent: text,
			}
		}

		rps.WriteMessage(&c.conn, m)
	}
}

// Handle message from server.
func (c *client) handleServerMessage() {
	for {
		buffer := make([]byte, 100)
		n, err := c.conn.Read(buffer)
		if err != nil {
			// disconnected from server
			fmt.Println("Disconnected from server")
			close(c.closeChan)
			return
		}
		m := rps.Message{}
		json.Unmarshal(buffer[:n], &m)

		switch m.MsgType {
		case rps.MsgConnected:
			c.state = clientConnected
			c.id, err = strconv.Atoi(m.MsgContent)
			if err != nil {
				fmt.Println("Unable to process player id")
			}
			fmt.Println("Successfully connected to server")
		case rps.MsgOponent:
			c.state = clientInGame
			fmt.Println("Found an oponent")
		case rps.MsgWaitMove:
			fmt.Println("Waiting for oponent")
		case rps.MsgGameEnd:
			c.state = clientConnected
			fmt.Println("Game has ended")
		default:
			fmt.Println("Unrecognized server message")
		}
	}
}
