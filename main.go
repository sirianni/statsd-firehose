package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	dataDogStatsd "github.com/DataDog/datadog-go/statsd"
	"github.com/stvp/clock"
)

var (
	// Settings
	statsdUrl        = flag.String("statsd", "127.0.0.1:8125", "Statsd URL")
	statsdPacketSize = flag.Int("packetsize", 512, "UDP packet size for metrics sent to statsd")

	useRandom = flag.Bool("random", false, "Use random values")

	gaugeCount    = flag.Int("gaugecount", 0, "Number of individual gauges to run")
	gaugeInterval = flag.Int("gaugeinterval", 1, "Gauge update interval, in seconds")
	gaugeFreq     = flag.Int("gaugefreq", 1, "How many times to update each individual gauge per interval")

	counterCount    = flag.Int("countcount", 0, "Number of individual counters to run")
	counterInterval = flag.Int("countinterval", 1, "Count update interval, in seconds")
	counterFreq     = flag.Int("countfreq", 1, "How many times to update each individual count per interval")

	distCount    = flag.Int("distcount", 0, "Number of individual distributions to run")
	distInterval = flag.Int("distinterval", 1, "Distribution update interval, in seconds")
	distFreq     = flag.Int("distfreq", 1, "How many times to update each individual distribution per interval")

	histCount    = flag.Int("histcount", 0, "Number of individual histograms to run")
	histInterval = flag.Int("histinterval", 1, "Histogram update interval, in seconds")
	histFreq     = flag.Int("histfreq", 1, "How many times to update each individual histogram per interval")

	verbose = flag.Bool("verbose", false, "Verbose print")
	tags    = flag.String("tags", "source:firehose", "Comma-separated list of tags to send with each metrics")

	namespace = flag.String("namespace", "firehose", "Namespace for firehose metrics")

	tagsArr []string

	gaugesUpdated   int64
	countersUpdated int64
	distUpdated     int64
	histUpdated     int64

	lastGauge   int64
	lastCounter int64
	lastDist    int64
	lastHist    int64
	// Globals
	client *dataDogStatsd.Client

	getInt = func() int64 {
		return 1
	}
	getFloat = func() float64 {
		return 0.5
	}
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
	if *useRandom {
		getInt = func() int64 {
			return int64(rand.Intn(10))
		}
		getFloat = func() float64 {
			return rand.NormFloat64()
		}
	}
	tagsArr = strings.Split(*tags, ",")
	log.SetOutput(os.Stdout)
}

func runGauges(count, freq int, interval time.Duration) {
	c, err := clock.New(100*time.Millisecond, interval)
	if err != nil {
		panic(err)
	}
	for key := range keys(count) {
		for i := 0; i < freq; i++ {
			c.Add(fmt.Sprintf("%s:%d", key, i))
		}
	}
	c.Start()

	for key := range c.Channel {
		client.Gauge("g", getFloat(), append(tagsArr, "no:"+strings.Split(key, ":")[0]), 1)
		verbosePrint("gauge: ", key)
		atomic.AddInt64(&gaugesUpdated, 1)
	}
}

func runCounters(count, freq int, interval time.Duration) {
	c, err := clock.New(100*time.Millisecond, interval)
	if err != nil {
		panic(err)
	}
	for key := range keys(count) {
		for i := 0; i < freq; i++ {
			c.Add(fmt.Sprintf("%s:%d", key, i))
		}
	}
	c.Start()

	for key := range c.Channel {
		client.Count("c", getInt(), append(tagsArr, "no:"+strings.Split(key, ":")[0]), 1)
		verbosePrint("count: ", key)
		atomic.AddInt64(&countersUpdated, 1)
	}
}

func runDist(count, freq int, interval time.Duration) {
	c, err := clock.New(100*time.Millisecond, interval)
	if err != nil {
		panic(err)
	}
	for key := range keys(count) {
		for i := 0; i < freq; i++ {
			c.Add(fmt.Sprintf("%s:%d", key, i))
		}
	}
	c.Start()

	for key := range c.Channel {
		client.Distribution("d", getFloat(), append(tagsArr, "no:"+strings.Split(key, ":")[0]), 1)
		verbosePrint("dist: ", key)
		atomic.AddInt64(&distUpdated, 1)
	}
}

func runHist(count, freq int, interval time.Duration) {
	c, err := clock.New(100*time.Millisecond, interval)
	if err != nil {
		panic(err)
	}
	for key := range keys(count) {
		for i := 0; i < freq; i++ {
			c.Add(fmt.Sprintf("%s:%d", key, i))
		}
	}
	c.Start()

	for key := range c.Channel {
		client.Histogram("h", getFloat(), append(tagsArr, "no:"+strings.Split(key, ":")[0]), 1)
		atomic.AddInt64(&histUpdated, 1)
	}
}

func keys(count int) chan string {
	c := make(chan string)
	go func() {
		for i := 0; i < count; i++ {
			c <- fmt.Sprintf("%d", i)
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

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		onExit()
		os.Exit(1)
	}()

	// Logging
	go func() {
		for _ = range time.Tick(time.Second) {
			log.Printf("gauges updated: %d; diff: %d", gaugesUpdated, gaugesUpdated-lastGauge)
			log.Printf("counters updated: %d, diff: %d", countersUpdated, countersUpdated-lastCounter)
			log.Printf("dists updated: %d, diff: %d", distUpdated, distUpdated-lastDist)
			log.Printf("hists updated: %d, diff: %d", histUpdated, histUpdated-lastHist)
			lastGauge = gaugesUpdated
			lastCounter = countersUpdated
			lastDist = distUpdated
			lastHist = histUpdated
		}
	}()

	// Turn on the firehose
	go func() { runGauges(*gaugeCount, *gaugeFreq, time.Duration(*gaugeInterval)*time.Second) }()
	go func() { runCounters(*counterCount, *counterFreq, time.Duration(*counterInterval)*time.Second) }()
	go func() { runDist(*distCount, *distFreq, time.Duration(*distInterval)*time.Second) }()
	go func() { runHist(*histCount, *histFreq, time.Duration(*histInterval)*time.Second) }()

	// Wait for Ctrl-C
	<-make(chan bool)
}

func onExit() {
	log.Printf(":: FINAL TOTALS")
	log.Printf("gauges updated: %d", gaugesUpdated)
	log.Printf("counters updated: %d", countersUpdated)
	log.Printf("dists updated: %d", distUpdated)
	log.Printf("hists updated: %d", histUpdated)
}
