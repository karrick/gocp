package gocp

import (
	"errors"

	"github.com/karrick/gopool"
)

type Pool struct {
	pool gopool.Pool
}

func New(setters ...Configurator) (*Pool, error) {
	pc := &poolConfig{
		poolSize: gopool.DefaultPoolSize,
	}
	for _, setter := range setters {
		if err := setter(pc); err != nil {
			return nil, err
		}
	}
	if pc.address == "" {
		return nil, errors.New("cannot create pool with empty address")
	}
	pool, err := gopool.NewChanPool(
		gopool.PoolSize(pc.poolSize),
		gopool.Factory(func() (interface{}, error) {
			return NewClient(ClientAddress(pc.address))
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Pool{pool: pool}, nil
}

func (pool *Pool) Get() *Client {
	return pool.pool.Get().(*Client)
}

func (pool *Pool) Put(client *Client) {
	pool.pool.Put(client)
}
