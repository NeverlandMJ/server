package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/NeverlandMJ/http/pkg/server"
)

const header = "HTTP/1.1 200 OK\r\n" +
	"Content-Length: %s\r\n" +
	"Content-Type: %s\r\n" +
	"Connection: close\r\n" +
	"\r\n"

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

func execute(host, port string) (err error) {
	srv := server.NewServer(net.JoinHostPort(host, port))

	srv.Register("/", func(req *server.Request) {
		body := "Welcome to our website"
		const header = "HTTP/1.1 200 OK\r\n" +
			"Content-Length: %s\r\n" +
			"Content-Type: %s\r\n" +
			"Connection: close\r\n" +
			"\r\n"

		id := req.QueryParams["id"]
		log.Println(id)
		_, err = req.Conn.Write([]byte(fmt.Sprintf(header, strconv.Itoa(len(body)), "text/html") + body))

		if err != nil {
			log.Println(err)
		}
	})

	srv.Register("/category{categoryID}/partners/{pID}", func(req *server.Request) {
		body := "Welcome to our website"
		const header = "HTTP/1.1 200 OK\r\n" +
			"Content-Length: %s\r\n" +
			"Content-Type: %s\r\n" +
			"Connection: close\r\n" +
			"\r\n"

		categoryID := req.PathParams["categoryID"]
		log.Println(categoryID)

		pID := req.PathParams["pID"]
		log.Println(pID)
		log.Println(req.Headers)

		_, err = req.Conn.Write([]byte(fmt.Sprintf(header, strconv.Itoa(len(body)), "text/html") + body))

		if err != nil {
			log.Println(err)
		}
	})

	return srv.Start()
}
