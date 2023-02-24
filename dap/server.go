package dap

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

type Server struct {
	listener  net.Listener
	config    *Config
	sessionMu sync.Mutex
	sessions  []*Session
}

func StartServer(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	disconnectChan := make(chan struct{})

	server := NewServer(&Config{
		log:            DAPLogger(),
		Listener:       listener,
		DisconnectChan: disconnectChan,
		StopTriggered:  make(chan struct{}),
	})
	defer server.Stop()
	server.Run()
	waitForDisconnectSignal(disconnectChan)
}

func StartInstance(port int) *Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Start Server Error: %v\n", err)
		return nil
	}
	disconnectChan := make(chan struct{})
	server := NewServer(&Config{
		log:            DAPLogger(),
		Listener:       listener,
		DisconnectChan: disconnectChan,
		StopTriggered:  make(chan struct{}),
	})
	go func(s *Server) {
		s.Run()
	}(server)
	return server
}

func (s *Server) Run() {
	if s.listener == nil {
		s.config.log.Fatal("Misconfigured server: no Listener is configured.")
		return
	}
	s.config.log.Info("server start")
	//go func() {
	for {
		conn, err := s.listener.Accept() // listener is closed in Stop()
		if err != nil {
			select {
			case <-s.config.StopTriggered:
				break
			default:
				s.config.log.Errorf("Error accepting client connection: %s\n", err)
				//s.config.triggerServerStop()
			}
			return
		}
		//if s.config.CheckLocalConnUser {
		//	if !sameuser.CanAccept(s.listener.Addr(), conn.LocalAddr(), conn.RemoteAddr()) {
		//		s.config.log.Error("Error accepting client connection: Only connections from the same user that started this instance of Delve are allowed to connect. See --only-same-user.")
		//		s.config.triggerServerStop()
		//		return
		//	}
		//}
		ip := ""
		if c, ok := conn.(net.Conn); ok {
			ip = c.RemoteAddr().String()
		}
		s.config.log.Warnf("client %s connect", ip)

		s.runSession(conn)
	}
	//}()
}

func (s *Server) Stop() {
	close(s.config.StopTriggered)

	if s.listener != nil {
		// If run goroutine is blocked on accept, this will unblock it.
		_ = s.listener.Close()
	}
	if s.sessions == nil {
		return
	}
	// If run goroutine is blocked on read, this will unblock it.
	for _, session := range s.sessions {
		session.Close()
	}

	s.config.log.Error("server stop")
}

func (s *Server) runSession(conn net.Conn) {
	s.sessionMu.Lock()
	session := NewSession(conn, s.config)
	s.sessions = append(s.sessions, session) // closed in Stop()
	s.sessionMu.Unlock()

	go session.ServeDAPCodec()
}

type Config struct {
	StopTriggered chan struct{}

	DisconnectChan chan<- struct{}
	Listener       net.Listener

	// log is used for structured logging.
	log *logrus.Entry
}

func NewServer(config *Config) *Server {
	return &Server{
		config:   config,
		listener: config.Listener,
		//StopTriggered: make(chan struct{}),
	}
}

// waitForDisconnectSignal is a blocking function that waits for either
// a SIGINT (Ctrl-C) or SIGTERM (kill -15) OS signal or for disconnectChan
// to be closed by the server when the client disconnects.
// Note that in headless mode, the debugged process is foregrounded
// (to have control of the tty for debugging interactive programs),
// so SIGINT gets sent to the debuggee and not to delve.
func waitForDisconnectSignal(disconnectChan chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	if runtime.GOOS == "windows" {
		// On windows Ctrl-C sent to inferior process is delivered
		// as SIGINT to delve. Ignore it instead of stopping the server
		// in order to be able to debug signal handlers.
		go func() {
			for range ch {
			}
		}()
		<-disconnectChan
	} else {
		select {
		case <-ch:
		case <-disconnectChan:
		}
	}
}
