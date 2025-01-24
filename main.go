package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type UnaggregatedFlowModel struct {
	DeviceProduct string `json:"DeviceProduct,omitempty"`
	StartTime     string `json:"StartTime,omitempty"`
}

func parseTime(s string) time.Time {
	for _, f := range []string{time.RFC3339Nano, time.RFC3339} {
		t, err := time.Parse(f, s)
		if err == nil {
			return t
		}
	}
	panic("can't parse " + s)
}

var password = flag.String("c", "", "connection string")
var topic = flag.String("t", "", "topic")
var broker = flag.String("b", "", "broker")
var partition = flag.Int("p", 0, "partition")
var messages = flag.Int("n", 1, "number of messages")

func check[T any](t *T, what string) {
	if t == nil {
		panic(fmt.Sprintf("%v must be set", what))
	}
}

func main() {
	flag.Parse()

	check(password, "connection string (-c)")
	check(topic, "topic (-t)")
	check(broker, "broker (-b)")
	check(partition, "partition (-p)")

	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.User = "$ConnectionString"
	config.Net.SASL.Password = *password
	config.Net.TLS.Enable = true
	config.Net.WriteTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	//config.Version = sarama.V1_1_0_0
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: false,
		ClientAuth:         0,
	}
	config.Consumer.Return.Errors = true

	cl, err := sarama.NewClient([]string{*broker}, config)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	offset, err := cl.GetOffset(*topic, int32(*partition), sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	c, err := sarama.NewConsumerFromClient(cl)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	pc, err := c.ConsumePartition(*topic, int32(*partition), offset-int64(*messages))
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for er := range pc.Errors() {
			fmt.Printf("[ERR] consumer error: %v", er)
		}
	}()

	stat := NewStats()

	fmt.Printf("Checking last %v messages in partition %v of topic %v\n", *messages, *partition, *topic)
	read := 0
	for msg := range pc.Messages() {
		var f UnaggregatedFlowModel
		err := json.Unmarshal(msg.Value, &f)
		if err != nil {
			panic(err)
		}
		t := parseTime(f.StartTime)
		stat.process(t, f.DeviceProduct != "")
		read++
		if read%10000 == 0 {
			fmt.Print(".")
		}
		if read == *messages {
			break
		}
	}
	pc.Close()
	c.Close()
	cl.Close()

	fmt.Println("\n", stat.String())
	wg.Wait()
}
