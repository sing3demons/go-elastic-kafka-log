package logger

import (
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const Key = "logger"
const XSession = "X-Request-Id"

func LoggingMiddleware(logger ILogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Starting time request
		startTime := time.Now()
		// Processing request
		reqId := ctx.Writer.Header().Get(XSession)
		if reqId == "" {
			reqId = uuid.NewString()
			ctx.Writer.Header().Set(XSession, reqId)
		}
		l := logParentID(ctx, logger)

		ctx.Set(Key, l)
		ctx.Next()
		// End Time request
		endTime := time.Now()
		// Request method
		reqMethod := ctx.Request.Method
		// Request route
		path := ctx.Request.RequestURI
		// status code
		statusCode := ctx.Writer.Status()

		// Request host
		host := ctx.Request.Host
		// Request user agent

		bodySize := ctx.Writer.Size()
		// execution time
		latencyTime := endTime.Sub(startTime)

		headers := GetHeaders(ctx)

		logger.Info("HTTP::REQUEST", LoggerFields{
			"headers":       headers,
			"method":        reqMethod,
			"status":        statusCode,
			"latency":       latencyTime,
			"error":         ctx.Errors.ByType(gin.ErrorTypePrivate).String(),
			"request":       ctx.Request.PostForm.Encode(),
			"body_size":     bodySize,
			"host":          host,
			"protocol":      ctx.Request.Proto,
			"path":          path,
			"query":         ctx.Request.URL.RawQuery,
			"response_size": ctx.Writer.Size(),
			"ContentType":   ctx.ContentType(),
			"ContentLength": ctx.Request.ContentLength,
			"timezone":      time.Now().Location().String(),
			"ISOTime":       startTime,
			"UnixTime":      startTime.UnixNano(),
		})

		ctx.Next()
	}
}

func logParentID(c *gin.Context, logger ILogger) ILogger {
	xParent := c.Request.Header.Get("X-Parent-ID")
	if xParent == "" {
		xParent = uuid.NewString()
	}
	xSpan := uuid.NewString()

	reqId := c.Writer.Header().Get(XSession)
	return logger.With(zap.String("parent-id", xParent),
		zap.String("span-id", xSpan), zap.String("session", reqId))
}

func GetHeaders(ctx *gin.Context) map[string]any {
	// Request user agent
	userAgent := ctx.Request.UserAgent()
	platform := strings.Split(ctx.Request.Header.Get("sec-ch-ua"), ",")
	mobile := ctx.Request.Header.Get("sec-ch-ua-mobile")
	operatingSystem := ctx.Request.Header.Get("sec-ch-ua-platform")
	clientIP := ctx.ClientIP()
	reqId := ctx.Writer.Header().Get(XSession)

	xParent := ctx.Request.Header.Get("X-Parent-ID")
	xSpan := uuid.NewString()

	macIp := getMACAndIP()

	return map[string]any{
		"user_agent": userAgent,
		"Platform":   platform,
		"Mobile":     mobile,
		"OS":         operatingSystem,
		"client_ip":  clientIP,
		"session":    reqId,
		"remote_ip":  ctx.Request.RemoteAddr,
		"mac_ip":     macIp,
		"parent_id":  xParent,
		"span_id":    xSpan,
	}
}

func getMACAndIP() MacIP {
	interfaces, _ := net.Interfaces()
	macAddr := MacIP{}
	for _, iface := range interfaces {

		if iface.Name != "" {
			macAddr.InterfaceName = iface.Name
		}

		if iface.HardwareAddr != nil {
			macAddr.HardwareAddr = iface.HardwareAddr.String()
		}

		var ips []string
		addrs, _ := iface.Addrs()

		for _, addr := range addrs {
			ips = append(ips, addr.String())
		}

		if len(ips) > 0 {
			macAddr.IPs = ips
		}
	}

	return macAddr
}

type MacIP struct {
	InterfaceName string   `json:"interface_name"`
	HardwareAddr  string   `json:"hardware_addr"`
	IPs           []string `json:"ips"`
}
