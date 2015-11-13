package main

import (
	"log"

	"github.com/karrick/gocp"
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
