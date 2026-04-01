package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

const SocketName = "doot.sock"

var logger = log.New(os.Stderr, "[doot] ", log.Ltime)

// Packet is the wire format for all messages.
type Packet struct {
	Type   string          `json:"type"`
	Event  string          `json:"event,omitempty"`
	Method string          `json:"method,omitempty"`
	ID     string          `json:"id,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Handler handles an incoming request and returns a result or error.
type Handler func(params json.RawMessage) (any, error)

// Server is the Unix socket server that broadcasts events and handles requests.
type Server struct {
	socketPath string
	listener   net.Listener
	clients    sync.Map
	handlers   map[string]Handler
	mu         sync.RWMutex
}

func SocketPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, SocketName)
	}
	return fmt.Sprintf("/run/user/%d/%s", os.Getuid(), SocketName)
}

func New() *Server {
	return &Server{
		socketPath: SocketPath(),
		handlers:   make(map[string]Handler),
	}
}

// Handle registers a request handler for a method name.
func (s *Server) Handle(method string, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = h
}

// Broadcast sends an event to all connected clients.
func (s *Server) Broadcast(event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	pkt := Packet{Type: "event", Event: event, Data: payload}
	line, err := json.Marshal(pkt)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	count := 0
	s.clients.Range(func(key, _ any) bool {
		_, _ = key.(net.Conn).Write(line)
		count++
		return true
	})

	logger.Printf("← event  %-20s %s  (to %d client(s))", event, payload, count)
	return nil
}

// Start begins listening on the Unix socket.
func (s *Server) Start() error {
	os.Remove(s.socketPath)
	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.socketPath, err)
	}
	s.listener = l

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			addr := conn.RemoteAddr()
			logger.Printf("+ client connected (%v)", addr)
			s.clients.Store(conn, struct{}{})
			go s.handleConn(conn)
		}
	}()
	return nil
}

// Stop shuts down the server and removes the socket file.
func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.socketPath)
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		s.clients.Delete(conn)
		logger.Printf("- client disconnected (%v)", conn.RemoteAddr())
		conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var pkt Packet
		if err := json.Unmarshal(scanner.Bytes(), &pkt); err != nil {
			logger.Printf("  bad JSON from client: %v", err)
			continue
		}
		if pkt.Type == "request" {
			logger.Printf("→ request  id=%-12s method=%s  params=%s", pkt.ID, pkt.Method, pkt.Data)
			go s.handleRequest(conn, pkt)
		}
	}
}

func (s *Server) handleRequest(conn net.Conn, req Packet) {
	s.mu.RLock()
	handler, ok := s.handlers[req.Method]
	s.mu.RUnlock()

	resp := Packet{Type: "response", ID: req.ID}

	if !ok {
		resp.Error = fmt.Sprintf("unknown method: %s", req.Method)
	} else {
		result, err := handler(req.Data)
		if err != nil {
			resp.Error = err.Error()
		} else if result != nil {
			resp.Data, _ = json.Marshal(result)
		}
	}

	if resp.Error != "" {
		logger.Printf("← response id=%-12s error=%s", resp.ID, resp.Error)
	} else {
		logger.Printf("← response id=%-12s data=%s", resp.ID, resp.Data)
	}

	line, _ := json.Marshal(resp)
	line = append(line, '\n')
	_, _ = conn.Write(line)
}
