package server

import (
	"bytes"
	"io"
	"log"
	"net"
	//"net/http"
	"strings"
	"sync"
)

type Request struct {
	Conn net.Conn
	PathParms map[string]string
}

type HandlerFunction func(conn net.Conn)

type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunction
}

func NewServer(addr string) *Server {
	return &Server{
		addr:     addr,
		mu:       sync.RWMutex{},
		handlers: make(map[string]HandlerFunction),
	}
}

func (s *Server) Register(path string, handler HandlerFunction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()
	if err != nil {
		log.Print(err)
		return err
	}
	for{
		conn, err := listener.Accept()
		if err != nil{
			log.Print(err)
			continue
		}
		go s.handle(conn)
	}

	return nil
}

func (s *Server) handle(conn net.Conn) (err error) {
	defer func ()  {
		if cerr := conn.Close(); cerr != nil {
			if cerr == err{
				err=cerr
				return
			}
			log.Print(err)
		}
	}()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		log.Printf("%s", buf[:n])
		return nil
	}
	if err != nil{
		return err
	}
	log.Printf("%s", buf[:n])

	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	if requestLineEnd == -1 {
	  return 
	}
  
	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
	  return
	}
  
	path, version := parts[1], parts[2]
  
	if version != "HTTP/1.1" {
	  return nil
	}

	s.mu.RLock()
  
	getHandle := s.handlers[path]
  
	if getHandle == nil {
	  return
	}
  
	s.mu.RUnlock()
  
	getHandle(conn)
	return

	
}

