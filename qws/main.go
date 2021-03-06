package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/thbkrkr/go-utilz/http"
	"github.com/thbkrkr/qli/client"
	"github.com/thbkrkr/qli/qws/controllers"
	"golang.org/x/net/websocket"
)

var (
	name      = "qws"
	buildDate = "dev"
	gitCommit = "dev"
	port      = 4242

	q    *client.Qlient
	hubs = map[string]*hub{}
)

func main() {
	//log.SetLevel(log.DebugLevel)
	hostname, _ := os.Hostname()

	var err error
	q, err = client.NewClientFromEnv(fmt.Sprintf("qws-%s", hostname))
	if err != nil {
		log.WithError(err).Fatal("Fail to create qlient")
	}

	go http.API(name, buildDate, gitCommit, port, router)

	q.CloseOnSig()
}

func router(r *gin.Engine) {
	wsCtl := WsCtrl{Q: q}

	r.GET("/ws", func(c *gin.Context) {
		topic := c.Query("topic")
		if topic == "" {
			c.JSON(400, "Invalid topic")
			return
		}

		SubWs(topic)
		log.WithField("topic", topic).Debug("Start sub to ws")

		handler := websocket.Handler(func(ws *websocket.Conn) {
			log.WithField("topic", topic).Debug("Start WS")
			wsCtl.WsPub(ws, topic)
		})
		handler.ServeHTTP(c.Writer, c.Request)
	})

	r.POST("/pub/:topic", Pub)
	r.GET("/stream/:topic", wsCtl.RawStream)
	r.GET("/metrics/:topic", controllers.Metrics)

}

func SubWs(topic string) {
	var err error
	sub, err := q.SubOn(topic)
	if err != nil {
		log.WithField("topic", topic).WithError(err).Error("Fail to create sub qlient")
		return
	}

	go func(h *hub) {
		for msg := range sub {
			log.WithField("topic", h.topic).WithField("msg", string(msg)).
				Debug("Broadcast kafka msg to the hub")
			h.broadcast <- msg
		}
	}(getHub(topic))
}
