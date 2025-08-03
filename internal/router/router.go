package router

import (
	"ccards/internal/client"
	"ccards/pkg/config"
	"ccards/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Router struct {
	engine        *gin.Engine
	clientHandler *client.Handler
	config        *config.Config
	redisClient   *redis.Client
}

type RouterConfig struct {
	ClientHandler *client.Handler
	Config        *config.Config
	RedisClient   *redis.Client
}

func NewRouter(cfg RouterConfig) *Router {
	return &Router{
		engine:        gin.New(),
		clientHandler: cfg.ClientHandler,
		config:        cfg.Config,
		redisClient:   cfg.RedisClient,
	}
}

func (r *Router) Setup() *gin.Engine {
	r.engine.Use(gin.Logger())
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.CORS())

	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	admin := r.engine.Group("/admin")
	{
		admin.POST("/company/register", r.clientHandler.RegisterCompany)
	}

	authGroup := r.engine.Group("/auth/company")
	{
		authGroup.POST("/login", r.clientHandler.Login)
	}

	apiGroup := r.engine.Group("/api")
	clientAuthMiddleware := middleware.NewClientAuthMiddleware(r.config, r.redisClient)
	apiGroup.Use(clientAuthMiddleware.ClientAuth())
	{
		companyGroup := apiGroup.Group("/company")
		{
			companyGroup.GET("", r.clientHandler.GetCompany)
			companyGroup.POST("/upload-csv", r.clientHandler.UploadCardCSV)
			companyGroup.GET("/to-issue", r.clientHandler.GetCardsToIssue)
		}
	}

	return r.engine
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
