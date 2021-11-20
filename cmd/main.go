package main

import (
	
	"net"
	"strconv"

	"os"

	"github.com/NeverlandMJ/http/pkg/server"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

func execute(host string, port string) (err error) {
	srv := server.NewServer(net.JoinHostPort(host, port))
	
	srv.Register("/", func(conn net.Conn) {
		body := "Welcome to our web-site"

		err = Body(body, conn)
		if err != nil {
			return
		}
	})
	srv.Register("/about", func(conn net.Conn) {
		body := "About Golang Academy"
		
		err = Body(body, conn)
		if err != nil {
			return
		}
	})
	return srv.Start()
}
func  Body(body string, conn net.Conn) (err error) {
	CRLF := "\r\n"
	_, err = conn.Write([]byte(
		"HTTP/1.1 200 Ok" + CRLF +
			"Content-Length: " + strconv.Itoa(len(body)) + CRLF +
			"Content-Type: text/html" + CRLF +
			"Connection: close" + CRLF +
			CRLF +
			body,
	))
	return err
}