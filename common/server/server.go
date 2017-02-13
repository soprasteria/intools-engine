package server

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/soprasteria/dockerapi"
	"github.com/soprasteria/intools-engine/common/websocket"
	"github.com/soprasteria/intools-engine/controllers"
	"github.com/soprasteria/intools-engine/groups"
	"github.com/soprasteria/intools-engine/intools"
	"gopkg.in/redis.v3"

	"github.com/gin-gonic/contrib/expvar"
)

type Daemon struct {
	Port   int
	Engine *gin.Engine
	level  string
}

func NewDaemon(port int, level string, dockerClient *dockerapi.Client, dockerHost string, redisClient *redis.Client) *Daemon {

	var engine *gin.Engine
	if level == string(gin.DebugMode) {
		log.Debug("Initializing daemon in debug mode")
		gin.SetMode(gin.DebugMode)
		engine = gin.Default()
		engine.LoadHTMLFiles("index.html")
		engine.GET("/", func(c *gin.Context) {
			c.HTML(200, "index.html", nil)
		})
	} else {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.Default()
	}
	engine.Use(gin.Recovery())
	intools.Engine = &intools.IntoolsEngineImpl{dockerClient, dockerHost, redisClient}
	daemon := &Daemon{port, engine, level}
	length := groups.GetGroupsLength()
	websocket.InitChannel(length)
	return daemon
}

func (d *Daemon) Run() {
	d.Engine.Run(fmt.Sprintf("0.0.0.0:%d", d.Port))
}

func (d *Daemon) SetRoutes(logPath string) {
	d.Engine.GET("/websocket", websocket.GetWS)
	d.Engine.GET("/debug/vars", expvar.Handler())
	d.Engine.GET("/groups", controllers.ControllerGetGroups)
	d.Engine.GET("/logs", func(c *gin.Context) { controllers.GetLogs(c, logPath) })

	allGroupRouter := d.Engine.Group("/groups/")
	{
		allGroupRouter.GET("", controllers.ControllerGetGroups)

		oneGroupRouter := allGroupRouter.Group("/:group")
		{
			oneGroupRouter.GET("", controllers.ControllerGetGroup)
			oneGroupRouter.POST("", controllers.ControllerPostGroup)
			oneGroupRouter.DELETE("", controllers.ControllerDeleteGroup)

			oneGroupConnectorRouter := oneGroupRouter.Group("/connectors")
			{
				oneGroupConnectorRouter.GET("", controllers.ControllerGetConnectors)
				oneGroupConnectorRouter.GET("/:connector", controllers.ControllerGetConnector)
				oneGroupConnectorRouter.POST("/:connector", controllers.ControllerCreateConnector)
				oneGroupConnectorRouter.DELETE("/:connector", controllers.ControllerDeleteConnector)
				oneGroupConnectorRouter.GET("/:connector/refresh", controllers.ControllerExecConnector)
				oneGroupConnectorRouter.GET("/:connector/result", controllers.ControllerGetConnectorResult)
				oneGroupConnectorRouter.GET("/:connector/exec", controllers.ControllerGetConnectorExecutor)
			}
		}
	}

}
