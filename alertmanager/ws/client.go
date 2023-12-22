package ws

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/gorilla/websocket"
)

// Client -
type Client struct {
	id     string
	socket *websocket.Conn
	server *Server
}

// NewClient - 
func NewClient(id string, conn *websocket.Conn, server *Server) *Client {
	return &Client{
		id:     id,
		socket: conn,
		server: server,
	}
}

// Close -
func (c *Client) Close() error {
	return c.socket.Close()
}

// Read - 
func (c *Client) Read() {
	defer func() {
		c.server.DelClient(c)
		c.Close()
	}()
	for {
		msgType, data, err := c.socket.ReadMessage()
		fmt.Println("websocket read data", msgType, data, err)
		if err != nil {
			return
		}
		if msgType == websocket.CloseMessage {
			return
		}
	}
}

// Write -
func (c *Client) Write(msgType int, data []byte) error {
	if err := c.socket.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return errors.WithStack(err)
	}
	if err := c.socket.WriteMessage(msgType, data); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
