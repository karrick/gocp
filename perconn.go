package gocp

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const DefaultJobQueueSize = 10

// DefaultRetryMin is the default minimum amount of time the client will wait to reconnect to
// a remote host if the connection drops.
const DefaultRetryMin = time.Second

// DefaultRetryMax is the default maximum amount of time the client will wait to reconnect to
// a remote host if the connection drops.
const DefaultRetryMax = time.Minute

// DefaultDialTimeout determines how long the client waits for a connection before timeout.
const DefaultDialTimeout = 5 * time.Second

// Client represents a proxying connection to an RRD reader on another host.  It attempts to
// reconnect to other host when operations on connection causes errors.
type Client struct {
	address     string
	retryMin    time.Duration
	retryMax    time.Duration
	dialTimeout time.Duration
	jobs        chan *rillJob
}

// ClientConfigurator is a function that modifies a Client structure during instantiation.
type ClientConfigurator func(*Client) error

// ClientAddress changes the network address used by a Client.
func ClientAddress(address string) ClientConfigurator {
	return func(c *Client) error {
		c.address = address
		return nil
	}
}

// ClientRetryMin controls the minimum amount of time a Client will wait between connection attempts
// to the remote host.
func ClientRetryMin(duration time.Duration) ClientConfigurator {
	return func(c *Client) error {
		c.retryMin = duration
		return nil
	}
}

// ClientRetryMax controls the maximum amount of time a Client will wait between connection attempts
// to the remote host.
func ClientRetryMax(duration time.Duration) ClientConfigurator {
	return func(c *Client) error {
		c.retryMax = duration
		return nil
	}
}

func ClientDialTimeout(duration time.Duration) ClientConfigurator {
	return func(c *Client) error {
		c.dialTimeout = duration
		return nil
	}
}

// NewClient returns a Client structure to be used as a proxy to issue queries to and receive
// response data from another host.
//
//	func example() {
//		client, err := gorrd.NewClient(gorrd.ClientAddress(":2000"),
//			gorrd.ClientRetryMin(100*time.Millisecond),
//			gorrd.ClientRetryMax(time.Second))
//		if err != nil {
//			panic(err)
//		}
//
//		// later on...
//
//		query := &gorrd.Query{
//			Start:                 1446760350,
//			End:                   1446760710,
//			Step:                  60,
//			Pathname:              "/some/path/to/file.rrd",
//		}
//		response := client.RequestWithTimeout(query, 5*time.Second)
//		_ = response
//	}
func NewClient(setters ...ClientConfigurator) (*Client, error) {
	client := &Client{
		dialTimeout: DefaultDialTimeout,
		retryMin:    DefaultRetryMin,
		retryMax:    DefaultRetryMax,
		jobs:        make(chan *rillJob, DefaultJobQueueSize),
	}
	for _, setter := range setters {
		if err := setter(client); err != nil {
			return nil, err
		}
	}
	if client.retryMin == 0 {
		return nil, fmt.Errorf("cannot create client with retry: %d", client.retryMin)
	}
	if client.retryMax == 0 {
		return nil, fmt.Errorf("cannot create client with retry: %d", client.retryMax)
	}
	if client.retryMax < client.retryMin {
		return nil, fmt.Errorf("cannot create client with retry max (%d) less than retry min (%d)", client.retryMax, client.retryMin)
	}
	if client.address == "" {
		return nil, fmt.Errorf("cannot create client with address: %q", client.address)
	}
	go func() {
		retry := client.retryMin
		for {
			conn, err := net.DialTimeout("tcp", client.address, client.dialTimeout)
			if err != nil {
				log.Printf("[WARNING] cannot connect: %s", err)
				time.Sleep(retry)
				retry *= 2
				if retry > client.retryMax {
					retry = client.retryMax
				}
				continue
			}

			err = client.proxy(conn) // doesn't return until err
			log.Printf("[WARNING] cannot proxy requests from %s: %s", client.address, err)

			retry = client.retryMin
			time.Sleep(retry)
		}
	}()
	return client, nil
}

func (client *Client) proxy(rwc io.ReadWriteCloser) error {
	defer rwc.Close()
	var err error
	for job := range client.jobs {
		switch job.op {
		case _read:
			n, err := rwc.Read(job.data)
			job.results <- rillResult{n, err}
			if err != nil {
				return err
			}
		case _write:
			n, err := rwc.Write(job.data)
			job.results <- rillResult{n, err}
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (client *Client) Read(data []byte) (int, error) {
	job := newRillJob(_read, make([]byte, len(data)))
	client.jobs <- job

	result := <-job.results
	copy(data, job.data)
	return result.n, result.err
}

func (client *Client) Write(data []byte) (int, error) {
	job := newRillJob(_write, data)
	client.jobs <- job

	result := <-job.results
	return result.n, result.err
}
