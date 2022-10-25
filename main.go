package main

import (
	"flag"
	"fmt"
	dataDogStatsd "github.com/DataDog/datadog-go/statsd"
	"github.com/stvp/clock"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	// Settings
	statsdUrl        = flag.String("statsd", "127.0.0.1:8125", "Statsd URL")
	statsdPacketSize = flag.Int("packetsize", 512, "UDP packet size for metrics sent to statsd")

	gaugeCount    = flag.Int("gaugecount", 50000, "Number of individual gauges to run")
	gaugeInterval = flag.Int("gaugeinterval", 60, "Gauge update interval, in seconds")

	counterCount    = flag.Int("countcount", 50000, "Number of individual counters to run")
	counterInterval = flag.Int("countinterval", 60, "Gauge update interval, in seconds")

	verbose = flag.Bool("verbose", false, "Verbose print")
	tags    = flag.String("tags", "source:firehose", "Comma-separated list of tags to send with each metrics")

	namespace = flag.String("namespace", "firehose", "Namespace for firehose metrics")

	tagsArr []string

	// Statistics
	gaugesUpdated   = 0
	countersUpdated = 0

	// Globals
	client *dataDogStatsd.Client
)

func setup() {
	flag.Parse()
	var err error
	client, err = dataDogStatsd.New(
		*statsdUrl,
		dataDogStatsd.WithMaxBytesPerPayload(*statsdPacketSize),
		dataDogStatsd.WithNamespace(*namespace),
		dataDogStatsd.WithoutTelemetry(),
	)
	if err != nil {
		log.Println("failed to create statsd client", err.Error())
		panic(err)
	}
	tagsArr = strings.Split(*tags, ",")
	log.SetOutput(os.Stdout)
}

func runGauges(count int, interval time.Duration) {
	c, err := clock.New(100*time.Millisecond, interval)
	if err != nil {
		panic(err)
	}
	for key := range keys("g", count) {
		c.Add(key)
	}
	c.Start()

	for key := range c.Channel {
		client.Gauge(key, rand.NormFloat64(), tagsArr, 1)
		verbosePrint("gauge: ", key)
		gaugesUpdated++
	}
}

func runCounters(count int, interval time.Duration) {
	c, err := clock.New(100*time.Millisecond, interval)
	if err != nil {
		panic(err)
	}
	for key := range keys("c", count) {
		c.Add(key)
	}
	c.Start()

	for key := range c.Channel {
		client.Count(key, 1, tagsArr, 1)
		verbosePrint("count: ", key)
		countersUpdated++
	}
}

func keys(prefix string, count int) chan string {
	c := make(chan string)
	go func() {
		for i := 0; i < count; i++ {
			c <- fmt.Sprintf("%s.%X", prefix, i)
		}
		close(c)
	}()
	return c
}

func verbosePrint(v ...any) {
	if *verbose {
		log.Println(v...)
	}
}

func main() {
	setup()
	log.Println("using the following config:", "statsdUrl", *statsdUrl, "statsdPacketSize", *statsdPacketSize, "gaugeCount", *gaugeCount, "gaugeInterval", *gaugeInterval, "counterCount", *counterCount, "counterInterval", *counterInterval, "verbose", *verbose)
	// Logging
	go func() {
		for _ = range time.Tick(time.Second) {
			log.Printf("gauges updated: %d", gaugesUpdated)
			log.Printf("counters updated: %d", countersUpdated)
		}
	}()

	// Turn on the firehose
	go func() { runGauges(*gaugeCount, time.Duration(*gaugeInterval)*time.Second) }()
	go func() { runCounters(*counterCount, time.Duration(*counterInterval)*time.Second) }()

	// Wait for Ctrl-C
	<-make(chan bool)
}
