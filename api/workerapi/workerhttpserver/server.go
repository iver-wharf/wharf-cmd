package workerhttpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-core/pkg/cacertutil"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

var cfg config.WorkerServerConfig

type module interface {
	register(g *gin.RouterGroup)
}

func init() {
	c, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Failed to read config:", err)
		os.Exit(1)
	}
	cfg = c.Worker.Server
}

// Server lets us start an HTTP server asynchronously.
type Server struct {
	onServeErrorHandler func(error)

	srv          *http.Server
	workerServer *workerServer
	isRunning    bool
}

// NewServer creates a new server that can be started by calling Start.
func NewServer(builder worker.Builder) *Server {
	return &Server{
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

	applyTLSConfig()
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
		Addr:    cfg.HTTP.BindAddress,
		Handler: r,
	}

	go func() {
		log.Debug().WithString("certsFile", cfg.CA.CertsFile).Message("")
		log.Debug().WithString("certKeyFile", cfg.CA.CertKeyFile).Message("")
		s.isRunning = true
		if err := s.srv.ListenAndServeTLS(cfg.CA.CertsFile, cfg.CA.CertKeyFile); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info().Message("Closed server.")
			} else if s.onServeErrorHandler != nil {
				s.onServeErrorHandler(err)
			}
		}
	}()
}

func applyTLSConfig() {
	if cfg.CA.CertsFile != "" {
		client, err := cacertutil.NewHTTPClientWithCerts(cfg.CA.CertsFile)
		if err != nil {
			log.Error().WithError(err).Message("Failed to get net/http.Client with certs")
			os.Exit(1)
		}
		http.DefaultClient = client
	}
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
	if len(cfg.HTTP.CORS.AllowOrigins) > 0 {
		log.Info().
			WithStringf("origin", "%v", cfg.HTTP.CORS.AllowOrigins).
			Message("Allowing origins in CORS.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = cfg.HTTP.CORS.AllowOrigins
		corsConfig.AddAllowHeaders("Authorization")
		corsConfig.AllowCredentials = true
		r.Use(cors.New(corsConfig))
	} else if cfg.HTTP.CORS.AllowAllOrigins {
		log.Info().Message("Allowing all origins in CORS.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		r.Use(cors.New(corsConfig))
	}
}
