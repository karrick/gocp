package gocp

import (
	"errors"
	"io"

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
			return conn.(io.ReadWriteCloser).Close()
		}),
		gopool.Factory(func() (interface{}, error) {
			if pc.printer != nil {
				return goperconn.New(goperconn.Address(pc.address), goperconn.Logger(pc.printer))
			}
			return goperconn.New(goperconn.Address(pc.address))
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Pool{pool: pool}, nil
}

// Close will close the pool, which invokes Close on all connections currently in the pool.
func (pool *Pool) Close() error {
	return pool.pool.Close()
}

// Get acquires a connection resource from the pool. Note that the Pool will eventually call Close
// on all network connections when the Pool's Close method is invoked. Therefore, the client is
// expected to *not* call Close on the object return by the Get method.
//
//    func codeWillCausePanic(pool *gocp.Pool) {
//        conn := pool.Get()
//
//        // ERROR: The next function that attempts to use this network connection will receive an error
//        // because it is attempting to perform I/O on a closed stream.  Additionally, when the Pool code
//        // attempts to recover from the error, it will invoke the Close method on the connection a second
//        // time, causing a runtime panic.  DON'T DO THIS!
//        defer conn.Close() // <--- should be `defer pool.Put(conn)`
//
//        _, err := conn.Write([]byte("hello, world"))
//        if err != nil {
//            log.Fatal(err)
//        }
//    }
func (pool *Pool) Get() io.ReadWriteCloser {
	return pool.pool.Get().(io.ReadWriteCloser)
}

// Put releases a connection resource to the pool.
func (pool *Pool) Put(conn io.ReadWriteCloser) {
	pool.pool.Put(conn)
}
