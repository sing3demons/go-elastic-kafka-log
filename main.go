package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type UserAction struct {
	UserId string `json:"userId"`
	Action string `json:"action"`
}

const url = "http://localhost:8083/connectors"

func createConnector() (any, error) {
	method := "POST"
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
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(string(body))

	return string(body), nil

}

func HealthCheck() {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
		// createConnector()
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if string(b) == "[]" {
		fmt.Println("Create connector")
		createConnector()
	} else {
		fmt.Println(string(b))
		fmt.Println("Connector already exists")
	}
}

func main() {

	producer, err := sarama.NewSyncProducer([]string{"localhost:29092"}, nil)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		resp, err := http.Get("http://localhost:8083/connectors")
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})

			return
		}

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})

			return
		}
		c.JSON(200, gin.H{
			"connectors": string(b),
		})
	})

	r.POST("topic", func(c *gin.Context) {
		r, err := createConnector()

		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})

			return
		}

		c.JSON(200, gin.H{
			"message": r,
		})

	})

	r.POST("/user-action", func(c *gin.Context) {
		var body UserAction
		c.BindJSON(&body)
		// body := UserAction{
		// 	UserId: "2",
		// 	Action: "login",
		// }

		jsonBody, _ := json.Marshal(body)

		msg := &sarama.ProducerMessage{
			Topic: "example-topic",
			Value: sarama.StringEncoder(jsonBody),
		}

		p, o, err := producer.SendMessage(msg)

		if err != nil {

			c.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": fmt.Sprintf("Partition: %d, Offset: %d", p, o),
		})
	})

	r.Run() // listen and serve on
}
