package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sing3demons/test-kafka-log/logger"
	"github.com/sing3demons/test-kafka-log/mlog"
	"github.com/sing3demons/test-kafka-log/router"
)

const url = "http://localhost:8083/connectors"

type Connector struct {
	Name   string `json:"name"`
	Config Config `json:"config"`
}

type Config struct {
	ConnectorClass              string `json:"connector.class"`
	TasksMax                    string `json:"tasks.max"`
	Topics                      string `json:"topics"`
	KeyIgnore                   string `json:"key.ignore"`
	SchemaIgnore                string `json:"schema.ignore"`
	ConnectionUrl               string `json:"connection.url"`
	TypeName                    string `json:"type.name"`
	Name                        string `json:"name"`
	ValueConverter              string `json:"value.converter"`
	ValueConverterSchemasEnable string `json:"value.converter.schemas.enable"`
}

type Event struct {
	Timestamp string `json:"@timestamp"`
	Header    Header `json:"header"`
	Body      any    `json:"body"`
}

type Header struct {
	Type      string `json:"type"`
	EventName string `json:"eventName"`
	SessionId string `json:"sessionId"`
	UserId    string `json:"userId"`
}

func init() {
	if os.Getenv("GIN_MODE") != gin.ReleaseMode {
		godotenv.Load(".env.dev")
	}
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}

	gin.SetMode(mode)
}

func main() {
	log := logger.NewLogger()
	producer, err := NewSyncProducer(strings.Split(os.Getenv("KAFKA_BROKERS"), ","), log)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	r := router.NewMicroservice(log)
	r.GET("/topics", func(c router.IContext) {
		l := mlog.L(c)
		result, err := HttpGetClient[[]string](url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		l.Info("result", logger.LoggerFields{
			"result": result,
		})

		c.JSON(http.StatusOK, result)
	})

	r.GET("/topic", func(c router.IContext) {
		l := mlog.L(c)
		results, err := HttpGetClient[[]string](url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		topics := strings.Split(os.Getenv("TOPIC_NAMES"), ",")
		l.Info("result", logger.LoggerFields{
			"result": results,
			"topics": topics,
		})

		connector, err := CreateConnector()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}

		c.JSON(http.StatusOK, connector)
	})

	r.POST("/logging", func(c router.IContext) {
		l := mlog.L(c)

		var req any
		if err := c.ReadBodyJSON(&req); err != nil {
			c.Error(http.StatusBadRequest, "invalid body", err)
			return
		}

		body := Event{
			Timestamp: time.Now().Format(time.RFC3339),
			Header: Header{
				EventName: "eventLog",
				Type:      "log",
				SessionId: c.SessionId(),
			},
			Body: req,
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		msg := &sarama.ProducerMessage{
			Topic: "example-topic",
			Value: sarama.StringEncoder(jsonBody),
		}

		p, o, err := producer.SendMessage(msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "error",
			})
			return
		}

		l.Info("result", logger.LoggerFields{
			"result": fmt.Sprintf("Partition: %d, Offset: %d", p, o),
			"topics": "example-topic",
			"header": body.Header,
			"body":   body.Body,
		})

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	r.StartHTTP()
}

func NewSyncProducer(kafkaBrokers []string, log logger.ILogger) (sarama.SyncProducer, error) {

	producer, err := sarama.NewSyncProducer(kafkaBrokers, nil)
	if err != nil {
		log.Error("NewSyncProducer", logger.LoggerFields{
			"error":  err.Error(),
			"topics": kafkaBrokers,
		})
		return nil, err
	}

	log.Info("NewSyncProducer", logger.LoggerFields{
		"topics": kafkaBrokers,
	})

	return producer, nil
}

func CreateConnector() (Connector, error) {
	payload := strings.NewReader(`
  {
	  "name": "elasticsearch-sink",
	  "config": {
		  "connector.class": "io.confluent.connect.elasticsearch.ElasticsearchSinkConnector",
		  "tasks.max": "1",
		  "topics": "example-topic",
		  "key.ignore": "true",
		  "schema.ignore": "true",
		  "connection.url": "http://elastic:9200",
		  "type.name": "_doc",
		  "name": "elasticsearch-sink",
		  "value.converter": "org.apache.kafka.connect.json.JsonConverter",
		  "value.converter.schemas.enable": "false"
	  }
  }`)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)

	if err != nil {
		fmt.Println(err)
		return Connector{}, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return Connector{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return Connector{}, err
	}
	var result Connector
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println(err)
		return Connector{}, err
	}

	return result, nil

}
