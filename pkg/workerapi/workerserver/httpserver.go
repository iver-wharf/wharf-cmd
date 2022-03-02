package workerserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/cacertutil"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

type module interface {
	register(g *gin.RouterGroup)
}

type httpServer struct {
	bindAddress  string
	srv          *http.Server
	workerServer *workerHTTPServer
	isRunning    bool

	onServeErrorHandler func(error)
}

// NewHTTPServer creates a new HTTP server that can be started by calling Start.
func NewHTTPServer(bindAddress string, buildStepLister wharfyml.BuildStepLister) Server {
	return &httpServer{
		bindAddress:  bindAddress,
		workerServer: newWorkerHTTPServer(buildStepLister),
	}
}

func (s *httpServer) SetOnServeErrorHandler(onServeErrorHandler func(error)) {
	s.onServeErrorHandler = onServeErrorHandler
}

func (s *httpServer) Serve() error {
	if err := s.ForceStop(); err != nil {
		return err
	}

	if err := applyHTTPClient(); err != nil {
		return err
	}

	r := gin.New()
	applyGinHandlers(r)
	applyCORSConfig(r)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	s.registerModules(g)
	return s.serve(r)
}

func (s *httpServer) GracefulStop() error {
	defer func() {
		s.isRunning = false
	}()
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(context.Background())
}

func (s *httpServer) ForceStop() error {
	defer func() {
		s.isRunning = false
	}()
	if s.srv == nil {
		return nil
	}
	return s.srv.Close()
}

func (s *httpServer) IsRunning() bool {
	return s.srv != nil && s.isRunning
}

func (s *httpServer) registerModules(r *gin.RouterGroup) {
	modules := []module{
		&buildModule{s.workerServer},
		&artifactModule{s.workerServer},
	}
	for _, module := range modules {
		module.register(r)
	}
}

func (s *httpServer) serve(r *gin.Engine) error {
	s.srv = &http.Server{
		Addr:    s.bindAddress,
		Handler: r,
	}
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	go func() {
		log.Info().Messagef("Listening and serving HTTP on %s", s.srv.Addr)
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
	return nil
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
