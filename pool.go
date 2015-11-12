package gocp

import (
	"errors"

	"github.com/karrick/goperconn"
	"github.com/karrick/gopool"
)

// DefaultSize specifies the number of connections to maintain to a single host for a connection
// pool instance.
const DefaultSize = 5

// Pool maintains a free-list of connections to a single host in a pool.
type Pool struct {
	pool gopool.Pool
}

// New returns a new Pool structure.
func New(setters ...Configurator) (*Pool, error) {
	pc := &config{
		size: DefaultSize,
	}
	for _, setter := range setters {
		if err := setter(pc); err != nil {
			return nil, err
		}
	}
	if pc.address == "" {
		return nil, errors.New("cannot create pool with empty address")
	}
	pool, err := gopool.New(
		gopool.Size(pc.size),
		gopool.Close(func(conn interface{}) error {
			return conn.(*goperconn.Conn).Close()
		}),
		gopool.Factory(func() (interface{}, error) {
			return goperconn.New(goperconn.Address(pc.address))
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Pool{pool: pool}, nil
}

// Get acquires a connection resource from the pool.
func (pool *Pool) Get() *goperconn.Conn {
	return pool.pool.Get().(*goperconn.Conn)
}

// Put releases a connection resource to the pool.
func (pool *Pool) Put(conn *goperconn.Conn) {
	pool.pool.Put(conn)
}
