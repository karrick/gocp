# gocp

Go connection pool.

## Description

Go connection pool provides a pool of network connections to a single
remote endpoint.

The connection pool uses
[go persistent connections](http://github.com/karrick/goperconn) for
each network connection, which attempts to reconnect to the remote end
when the connection has an error.

## Usage Example

### Basic Example

Only the Address of the remote host is required. All other parameters
have reasonalbe defaults and may be elided.

```Go
    package main

    import (
        "log"
        gocp "gopkg.in/karrick/gocp.v1"
    )

    func main() {
        pool, err := gocp.New(gocp.Address("echo-server.example.com:7"))

        // later ...

        conn := pool.Get()
        defer pool.Put(conn)

        _, err = conn.Write([]byte("hello, world"))
        if err != nil {
            log.Fatal(err)
        }

        buf := make([]byte, 512)
        _, err = conn.Read(buf)
        if err != nil {
            log.Fatal(err)
        }
    }
```

### Logger

If there are connection errors, or an error that takes place during an
I/O operation, and you want to be notified of the error, you may pass
an object that has a `Print(...interface{})` method, and it will be
called with the error message.

```Go
    package main

    import (
        "log"
        gocp "gopkg.in/karrick/gocp.v1"
    )

    func main() {

        printer := log.New(os.Stderr, "WARNING: ", 0)

        pool, err := gocp.New(gocp.Address("echo-server.example.com:7"),
            gocp.Logger(printer))

        // later ...

        conn := pool.Get()
        defer pool.Put(conn)

        _, err = conn.Write([]byte("hello, world"))
        if err != nil {
            log.Fatal(err)
        }

        buf := make([]byte, 512)
        _, err = conn.Read(buf)
        if err != nil {
            log.Fatal(err)
        }
    }
```

### Size

If the desired pool size is different than the default, DefaultSize,
it can be changed by using the Size function.

```Go
    package main

    import (
        "log"
        gocp "gopkg.in/karrick/gocp.v1"
    )

    func main() {

        pool, err := gocp.New(gocp.Address("echo-server.example.com:7"),
            gocp.Size(10))

        // later ...

        conn := pool.Get()
        defer pool.Put(conn)

        _, err = conn.Write([]byte("hello, world"))
        if err != nil {
            log.Fatal(err)
        }

        buf := make([]byte, 512)
        _, err = conn.Read(buf)
        if err != nil {
            log.Fatal(err)
        }
    }
```

## Implementation Notes

* If you close the pool, only the connections in the pool at the time
  it is closed will be closed. Connections which have been acquired
  and not yet released will remain open.
