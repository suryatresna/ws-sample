package room

import "net"

type socketInterface interface {
	Add(conn net.Conn) (fd int, err error)
}

type Room struct {
	roomID     string
	socket     socketInterface
	mapRoomFds map[string][]int
}

func New(socket socketInterface) (*Room, error) {
	return &Room{
		socket:     socket,
		mapRoomFds: make(map[string][]int),
	}, nil
}

// Register is for add new room
func (r *Room) Register(roomID string) error {
	r.roomID = roomID
	return nil
}

// AddClient is add new client in room
func (r *Room) AddClient(roomID string, conn net.Conn) error {
	fd, err := r.socket.Add(conn)
	if err != nil {
		return err
	}
	r.mapRoomFds[roomID] = append(r.mapRoomFds[roomID], fd)
	return nil
}
