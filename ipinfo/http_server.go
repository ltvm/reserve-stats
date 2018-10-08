package ipinfo

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrInvalidIP error for invalid ip input
const ErrInvalidIP = "Invalid ip input"

// HTTPServer to serve endpoint
type HTTPServer struct {
	r     *gin.Engine
	l     *Locator
	port  int
	sugar *zap.SugaredLogger
}

// NewHTTPServer return an instance of HTTPServer
func NewHTTPServer(sugar *zap.SugaredLogger, dataDir string, port int) (*HTTPServer, error) {
	l, err := NewLocator(sugar, dataDir)
	if err != nil {
		return nil, err
	}
	return &HTTPServer{
		r:     gin.Default(),
		l:     l,
		port:  port,
		sugar: sugar,
	}, nil
}

func (h *HTTPServer) register() {
	h.r.GET("/ip/:ip", h.lookupIPCountry)
}

// Run start HTTPServer
func (h *HTTPServer) Run() error {
	h.register()
	port := fmt.Sprintf(":%d", h.port)
	return h.r.Run(port)
}

func (h *HTTPServer) lookupIPCountry(c *gin.Context) {
	ip := c.Param("ip")
	ipParsed := net.ParseIP(ip)
	if ipParsed == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": ErrInvalidIP,
			},
		)
		return
	}
	location, err := h.l.IPToCountry(ipParsed)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{
			"country": location,
		},
	)
}
