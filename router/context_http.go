package router

import (
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type IContext interface {
	QueryString(name string) string
	Param(key string) string

	JSON(code int, obj any)
	Body(obj any) error
	ReadBodyJSON(obj any) error
	Error(code int, msg string, err error)

	GetHeader() ([]sarama.RecordHeader, map[string]any)
	SetAuthorization(value string)
}

type HTTPContext struct {
	*Microservice
	*gin.Context
}

func NewContext(ms *Microservice, c *gin.Context) *HTTPContext {
	return &HTTPContext{ms, c}
}

func (c *HTTPContext) JSON(code int, obj any) {
	c.Context.JSON(code, obj)
}

func (ctx *HTTPContext) Body(obj any) error {
	err := ctx.Context.ShouldBind(&obj)
	if err != nil {
		return err
	}

	return nil
}
func (ctx *HTTPContext) ReadBodyJSON(obj any) error {
	err := ctx.Context.ShouldBindJSON(&obj)
	if err != nil {

		return err
	}

	return nil
}

func (ctx *HTTPContext) SetAuthorization(value string) {
	ctx.Context.Request.Header.Set("Authorization", "Bearer "+value)
}

func (c *HTTPContext) Error(code int, msg string, err error) {
	c.Context.JSON(code, map[string]any{
		"statusCode": code,
		"error":      err.Error(),
		"message":    msg,
	})
}

func (c *HTTPContext) QueryString(name string) string {
	return c.Context.Query(name)
}
func (c *HTTPContext) Param(key string) string {
	return c.Context.Param(key)
}

func (c *HTTPContext) GetHeader() ([]sarama.RecordHeader, map[string]any) {
	statusCode := strconv.Itoa(c.Writer.Status())
	authorization := c.Request.Header.Get("Authorization")
	headers := []sarama.RecordHeader{
		{Key: []byte("method"), Value: []byte(c.Request.Method)},
		{Key: []byte("status"), Value: []byte(statusCode)},
		{Key: []byte("client_ip"), Value: []byte(c.ClientIP())},
		{Key: []byte("request_id"), Value: []byte(c.Writer.Header().Get("X-Request-Id"))},
		{Key: []byte("remote_ip"), Value: []byte(c.Request.RemoteAddr)},
		{Key: []byte("user_id"), Value: []byte(c.Request.URL.User.Username())},
		{Key: []byte("user_agent"), Value: []byte(c.Request.UserAgent())},
		{Key: []byte("error"), Value: []byte(c.Errors.ByType(gin.ErrorTypePrivate).String())},
		{Key: []byte("request"), Value: []byte(c.Request.PostForm.Encode())},
		{Key: []byte("body_size"), Value: []byte(strconv.Itoa(c.Writer.Size()))},
		{Key: []byte("host"), Value: []byte(c.Request.Host)},
		{Key: []byte("protocol"), Value: []byte(c.Request.Proto)},
		{Key: []byte("path"), Value: []byte(c.Request.RequestURI)},
		{Key: []byte("timezone"), Value: []byte(time.Now().Location().String())},
		{Key: []byte("response_size"), Value: []byte(strconv.Itoa(c.Writer.Size()))},
		{Key: []byte("Authorization"), Value: []byte(authorization)},
	}

	header := map[string]any{
		"method":        c.Request.Method,
		"status":        statusCode,
		"client_ip":     c.ClientIP(),
		"request_id":    c.Writer.Header().Get("X-Request-Id"),
		"remote_ip":     c.Request.RemoteAddr,
		"user_id":       c.Request.URL.User.Username(),
		"user_agent":    c.Request.UserAgent(),
		"error":         c.Errors.ByType(gin.ErrorTypePrivate).String(),
		"request":       c.Request.PostForm.Encode(),
		"body_size":     c.Writer.Size(),
		"host":          c.Request.Host,
		"protocol":      c.Request.Proto,
		"path":          c.Request.RequestURI,
		"timezone":      time.Now().Location().String(),
		"response_size": c.Writer.Size(),
		"Authorization": authorization,
	}

	return headers, header
}
