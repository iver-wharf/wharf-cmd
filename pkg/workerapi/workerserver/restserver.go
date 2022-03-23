package workerserver

import (
	"net"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

type module interface {
	register(g *gin.RouterGroup)
}

type restServer struct {
	artifactOpener ArtifactFileOpener
}

func newRestServer(artifactOpener ArtifactFileOpener) *restServer {
	return &restServer{
		artifactOpener: artifactOpener,
	}
}

func serveHTTP(workerServer *server, s *restServer, listener net.Listener) error {
	r := gin.New()
	applyGinHandlers(r)
	applyCORSConfig(r)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	g.GET("/kill", func(c *gin.Context) {
		c.JSON(200, "Killing server")
		go func() {
			<-c.Request.Context().Done()
			workerServer.Close()
		}()
	})

	s.registerModules(g)
	return r.RunListener(listener)
}

func (s *restServer) registerModules(r *gin.RouterGroup) {
	modules := []module{
		&artifactModule{s.artifactOpener},
	}
	for _, module := range modules {
		module.register(r)
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
	log.Info().Message("Allowing all origins in CORS.")
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))
}
