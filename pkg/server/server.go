package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

//HandlerFunc ..
type HandlerFunc func(req *Request)

var (
	//ErrBadRequest ...
	ErrBadRequest = errors.New("Bad Request")
	//ErrMethodNotAlowed ..
	ErrMethodNotAlowed = errors.New("Method not Alowed")
	//ErrHTTPVersionNotValid ..
	ErrHTTPVersionNotValid = errors.New("Http version not valid")
)

//Server ..
type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

//Request ...
type Request struct {
	Conn        net.Conn
	QueryParams url.Values
	PathParams  map[string]string
	Headers     map[string]string
	Body        []byte
}

//NewServer ...
func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

//Register ...
func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

//Start ...
func (s *Server) Start() (err error) {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			err = cerr
			return
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go s.handle(conn)

	}

	//return nil
}

func (s *Server) handle(conn net.Conn) {

	defer conn.Close()

	buf := make([]byte, (1024 * 8))
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			log.Printf("%s", buf[:n])
		}
		if err != nil {
			log.Println(err)
			return
		}

		var req Request
		data := buf[:n]
		rLD := []byte{'\r', '\n'}
		rLE := bytes.Index(data, rLD)
		if rLE == -1 {
			log.Printf("Bad Request")
			return
		}

		// headers
		hLD := []byte{'\r', '\n', '\r', '\n'}
		hLE := bytes.Index(data, hLD)
		if rLE == -1 {
			return
		}

		headersLine := string(data[rLE:hLE])
		headers := strings.Split(headersLine, "\r\n")[1:]
		//headers = headers[1:]
		mp := make(map[string]string)
		for _, v := range headers {
			headerLine := strings.Split(v, ": ")
			mp[headerLine[0]] = headerLine[1]
		}

		req.Headers = mp

		// Body
		b := string(data[hLE:])
		b = strings.Trim(b, "\r\n")

		req.Body = []byte(b)

		reqLine := string(data[:rLE])
		parts := strings.Split(reqLine, " ")

		if len(parts) != 3 {
			log.Println(ErrBadRequest)
			return
		}
		//method, path, version := parts[0], parts[1], parts[2]
		path, version := parts[1], parts[2]
		if version != "HTTP/1.1" {
			log.Println(ErrHTTPVersionNotValid)
			return
		}

		decode, err := url.PathUnescape(path)
		if err != nil {
			log.Println(err)
			return
		}

		uri, err := url.ParseRequestURI(decode)
		if err != nil {
			log.Println(err)
			return
		}

		req.Conn = conn
		req.QueryParams = uri.Query()

		var handler = func(req *Request) { conn.Close() }

		s.mu.RLock()
		pParam, hr := s.checkPath(uri.Path)
		if hr != nil {
			handler = hr
			req.PathParams = pParam
		}
		s.mu.RUnlock()

		handler(&req)

	}

}

func (s *Server) checkPath(path string) (map[string]string, HandlerFunc) {

	strRoutes := make([]string, len(s.handlers))
	i := 0
	for k := range s.handlers {
		strRoutes[i] = k
		i++
	}

	mp := make(map[string]string)

	for i := 0; i < len(strRoutes); i++ {
		flag := false
		route := strRoutes[i]
		partsRoute := strings.Split(route, "/")
		pRotes := strings.Split(path, "/")

		for j, v := range partsRoute {
			if v != "" {
				f := v[0:1]
				l := v[len(v)-1:]
				if f == "{" && l == "}" {
					mp[v[1:len(v)-1]] = pRotes[j]
					flag = true
				} else if pRotes[j] != v {

					strs := strings.Split(v, "{")
					if len(strs) > 0 {
						key := strs[1][:len(strs[1])-1]
						mp[key] = pRotes[j][len(strs[0]):]
						flag = true
					} else {
						flag = false
						break
					}
				}
				flag = true
			}
		}
		if flag {
			if hr, found := s.handlers[route]; found {
				return mp, hr
			}
			break
		}
	}

	return nil, nil

}
