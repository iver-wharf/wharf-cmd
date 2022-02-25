package workerhttpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-core/pkg/cacertutil"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

type module interface {
	register(g *gin.RouterGroup)
}

// Server lets us start an HTTP server asynchronously.
type Server struct {
	address      string
	port         string
	srv          *http.Server
	workerServer *workerServer
	isRunning    bool

	onServeErrorHandler func(error)
}

// NewServer creates a new server that can be started by calling Start.
func NewServer(address string, port string, builder worker.Builder) *Server {
	return &Server{
		address:      address,
		port:         port,
		workerServer: newWorkerServer(builder),
	}
}

// SetOnServeErrorHandler sets the handler to call when an error occurs during
// serving.
func (s *Server) SetOnServeErrorHandler(onServeErrorHandler func(error)) {
	s.onServeErrorHandler = onServeErrorHandler
}

// Serve starts the HTTP server in a goroutine.
//
// Also functions as a force-restart by calling ForceStop if the server is
// already running, followed by attempting to launch it again.
//
// To stop the server you may use GracefulStop or ForceStop.
func (s *Server) Serve() {
	s.ForceStop()

	if err := applyHTTPClient(); err != nil {
		s.onServeErrorHandler(err)
		return
	}

	r := gin.New()
	applyGinHandlers(r)
	applyCORSConfig(r)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	s.registerModules(g)
	s.serve(r)
}

// GracefulStop stops the HTTP server gracefully, blocking new connections
// and closing idle connections, then waiting until active ones have finished
// or the timeout duration has elapsed.
//
// If 0 is passed as the timeout, a default value of 30 seconds will be used.
func (s *Server) GracefulStop(timeout time.Duration) error {
	defer func() {
		s.isRunning = false
	}()
	if s.srv == nil {
		return nil
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// ForceStop forcefully stops the server.
func (s *Server) ForceStop() error {
	defer func() {
		s.isRunning = false
	}()
	if s.srv == nil {
		return nil
	}
	return s.srv.Close()
}

// IsRunning returns true if the server is currently running and processing
// requests.
func (s *Server) IsRunning() bool {
	return s.srv != nil && s.isRunning
}

func (s *Server) registerModules(r *gin.RouterGroup) {
	modules := []module{
		&buildModule{s.workerServer},
		&artifactModule{s.workerServer},
	}
	for _, module := range modules {
		module.register(r)
	}
}

func (s *Server) serve(r *gin.Engine) {
	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.address, s.port),
		Handler: r,
	}
	go func() {
		log.Info().Messagef("Listening and serving HTTP on %s", s.srv.Addr)
		ln, err := net.Listen("tcp", s.srv.Addr)
		if err != nil {
			s.onServeErrorHandler(err)
			return
		}
		s.isRunning = true
		if err := s.srv.Serve(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info().Message("Server closed.")
			} else if s.onServeErrorHandler != nil {
				s.onServeErrorHandler(err)
			}
		}
		s.isRunning = false
	}()
	for !s.isRunning {
		time.Sleep(1 * time.Millisecond)
	}
}

func applyHTTPClient() error {
	client, err := cacertutil.NewHTTPClientWithCerts("/etc/iver-wharf/wharf-cmd/localhost.crt")
	if err == nil {
		http.DefaultClient = client
	}
	return err
}

func applyGinHandlers(r *gin.Engine) {
	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)
}

func applyCORSConfig(r *gin.Engine) {
	log.Info().Message("Allowing all origins in CORS.")
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))
}
