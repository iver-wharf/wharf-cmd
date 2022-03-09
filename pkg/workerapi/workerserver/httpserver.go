package workerserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

type module interface {
	register(g *gin.RouterGroup)
}

type httpServer struct {
	bindAddress  string
	srv          *http.Server
	workerServer *workerHTTPServer
	serveMutex   *sync.Mutex
}

// NewHTTPServer creates a new HTTP server that can be started by calling Serve.
func NewHTTPServer(
	bindAddress string,
	stepLister StepLister,
	artifactLister ArtifactLister,
	artifactDownloader ArtifactDownloader) Server {
	return &httpServer{
		bindAddress:  bindAddress,
		workerServer: newWorkerHTTPServer(stepLister, artifactLister, artifactDownloader),
		serveMutex:   &sync.Mutex{},
	}
}

func (s *httpServer) GracefulStop() error {
	if s.srv == nil {
		return nil
	}
	err := s.srv.Shutdown(context.Background())
	s.srv = nil
	return err
}

func (s *httpServer) ForceStop() error {
	if s.srv == nil {
		return nil
	}
	err := s.srv.Close()
	s.srv = nil
	return err
}

func (s *httpServer) IsRunning() bool {
	return s.srv != nil
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

func (s *httpServer) Serve() error {
	if err := s.ForceStop(); err != nil {
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

func (s *httpServer) serve(r *gin.Engine) error {
	s.serveMutex.Lock()
	defer s.serveMutex.Unlock()
	s.srv = &http.Server{
		Addr:    s.bindAddress,
		Handler: r,
	}
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	err = s.srv.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		log.Info().Message("Server closed.")
		return nil
	}
	return err
}
