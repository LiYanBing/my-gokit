package grpc_tool

import (
	"io"
	"sync"

	"google.golang.org/grpc"
)

var (
	pool  sync.Map
	mutex sync.RWMutex
)

func init() {
	pool = sync.Map{}
}

type ConnCloser struct {
	closed bool
	sync.Mutex
	instance string
	conn     *grpc.ClientConn
}

func (c *ConnCloser) Close() error {
	conn, ok := pool.Load(c.instance)
	if !ok {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	if c.closed {
		return nil
	}

	pool.Delete(c.instance)
	if conn == c.conn && c.conn != nil {
		c.conn.Close()
		c.closed = true
	}
	return nil
}

func Get(instance string, opts ...grpc.DialOption) (*grpc.ClientConn, io.Closer, error) {
	mutex.Lock()
	defer mutex.Unlock()
	conn, ok := pool.Load(instance)
	if ok {
		clientConn := conn.(*grpc.ClientConn)
		return clientConn, &ConnCloser{
			instance: instance,
			conn:     clientConn,
		}, nil
	}

	newConn, err := grpc.Dial(instance, opts...)
	if err != nil {
		return nil, nil, err
	}

	pool.Store(instance, newConn)
	return newConn, &ConnCloser{
		instance: instance,
		conn:     newConn,
	}, nil
}
