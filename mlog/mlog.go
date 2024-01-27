package mlog

import (
	"github.com/sing3demons/test-kafka-log/logger"
	"github.com/sing3demons/test-kafka-log/router"
)

func L(c router.IContext) logger.ILogger {
	value, _ := c.Get(logger.Key)
	switch log := value.(type) {
	case logger.ILogger:
		return log
	default:
		return logger.NewLogger()
	}
}
