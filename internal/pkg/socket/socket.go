package socket

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type pollerInterface interface {
	Add(conn net.Conn) (fd int, err error)
	Remove(conn net.Conn) error
	Wait(count int) ([]net.Conn, error)
	GetMapConnections() map[int]net.Conn
}

type Socket struct {
	receive chan *Receive
	epoll   pollerInterface
	mu      sync.RWMutex
}

type Receive struct {
	Message []byte
	Op      ws.OpCode
	RoomID  string
}

func NewSocket(poller pollerInterface) *Socket {
	return &Socket{
		receive: make(chan *Receive),
		epoll:   poller,
		mu:      sync.RWMutex{},
	}
}

func (s *Socket) Upgrade(w http.ResponseWriter, r *http.Request) error {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}

	_, err = s.Add(conn)
	if err != nil {
		return err
	}

	return nil

}

func (s *Socket) Add(conn net.Conn) (fd int, err error) {
	if conn == nil {
		return 0, errors.New("connection fail to add")
	}

	return s.epoll.Add(conn)
}

func (s *Socket) Start() {
	for {
		connections, err := s.epoll.Wait(1000)
		if err != nil {
			log.Println("Fail to collecting connection")
			return
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}

			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				// handle error
				if errEpoll := s.epoll.Remove(conn); errEpoll != nil {
					log.Println("fail remove epoll ", errEpoll)
				}
				conn.Close()
				log.Println("[INFO] connection break ", err)
				break
			}

			recv := &Receive{
				Message: msg,
				Op:      op,
			}

			s.receive <- recv

			log.Println("[DEBUG] send message ", string(recv.Message))
		}
	}
}

func (s *Socket) Broadcast(ctx context.Context, msg []byte) error {
	recv := &Receive{
		Message: msg,
		Op:      ws.OpText,
	}

	s.receive <- recv

	return nil
}

func (s *Socket) Listen() {
	for {
		s.mu.RLock()
		conns := s.epoll.GetMapConnections()
		s.mu.RUnlock()
		select {
		case recv, ok := <-s.receive:
			if !ok {
				return
			}

			for _, conn := range conns {
				err := wsutil.WriteServerMessage(conn, ws.OpText, recv.Message)
				if err != nil {
					log.Println("[ERR] write message ", err, "; recv message: ", string(recv.Message))
				}
			}

			log.Println("[DEBUG] receive message ", string(recv.Message))
		default:
		}
	}
}
