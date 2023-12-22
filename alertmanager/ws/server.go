package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server -
type Server struct {
	clients *sync.Map 
	closeCh chan struct{}
}

// NewServer -
func NewServer() *Server {
	s := &Server{
		clients: new(sync.Map),
		closeCh: make(chan struct{}),
	}

	return s
}

// Start -
func (s *Server) Start() {
	go func() {
		defer func() {
			s.clients.Range(func(key, value interface{}) bool {
				client := value.(*Client)
				if err := client.Close(); err != nil {
					fmt.Println("websocket close client error", err)
				}
				return true
			})
		}()

		ticker := time.NewTicker(time.Second * 10)
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-s.closeCh:
				return
			case <-ticker.C:
				s.clients.Range(func(key, value interface{}) bool {
					client := value.(*Client)
					if err := client.Write(websocket.PingMessage, nil); err != nil {
						fmt.Println("websocket ping client error", err)

						s.DelClient(client)

						if err := client.Close(); err != nil {
							fmt.Println("websocket close client error", err)
						}
					}
					return true
				})
			}
		}
	}()
}

// Close -
func (s *Server) Close() {
	close(s.closeCh)
}

// AddClient -
func (s *Server) AddClient(client *Client) {
	s.clients.Store(client.id, client)
}

// DelClient -
func (s *Server) DelClient(client *Client) {
	s.clients.Delete(client.id)
}

// Broadcast - 
func (s *Server) Broadcast(data []byte) {
	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		if err := client.Write(websocket.TextMessage, data); err != nil {
			fmt.Println("websocket write data error", err)
		}
		return true
	})
}
