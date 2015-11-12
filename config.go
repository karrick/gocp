package gocp

import (
	"fmt"

	"github.com/karrick/goperconn"
)

type config struct {
	address string
	printer goperconn.Printer
	size    int
}

// Configurator is a function that modifies a pool configuration structure.
type Configurator func(*config) error

// Address specifies the network address (hostname:port) to use to create a socket to the remote
// host.
func Address(address string) Configurator {
	return func(pc *config) error {
		pc.address = address
		return nil
	}
}

// Logger specifies the logger to be used to log I/O errors.
func Logger(printer goperconn.Printer) Configurator {
	return func(pc *config) error {
		pc.printer = printer
		return nil
	}
}

// Size specifies the number of buffers to maintain in the pool.
func Size(size int) Configurator {
	return func(pc *config) error {
		if size <= 0 {
			return fmt.Errorf("pool size must be greater than 0: %d", size)
		}
		pc.size = size
		return nil
	}
}
