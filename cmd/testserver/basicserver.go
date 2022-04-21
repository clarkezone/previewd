// package testserver contains a dummy server implementation for testing metrics and logging
package testserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type cleanupfunc func()

//var requestsProcessed = promauto.NewCounter(prometheus.CounterOpts{
//	Name: "go_request_operations_total",
//	Help: "The total number of processed requests",
//})
//
//var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
//	Name:    "go_request_duration_seconds",
//	Help:    "Histogram for the duration in seconds.",
//	Buckets: []float64{1, 2, 5, 6, 10},
//},
//	[]string{"endpoint"},
//)

// BasicServer object
type BasicServer struct {
	httpserver *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

// Create BasicServer object and return
func CreateBasicServer() *BasicServer {
	bs := BasicServer{}
	return &bs
}

// Start listening for a connection
func (bs *BasicServer) StartListen(secret string) {
	log.Println("starting... basic server on :8090")

	//prometheus.MustRegister(requestDuration)

	bs.ctx, bs.cancel = context.WithCancel(context.Background())
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		//message := fmt.Sprintf("Hello World")
		//w.Write([]byte(message))
		//start a timer
		//start := time.Now()

		//measure the duration and log to prometheus
		//httpDuration := time.Since(start)
		//requestDuration.WithLabelValues("GET /").Observe(httpDuration.Seconds())

		//increment a counter for number of requests processed
		//requestsProcessed.Inc()
	})
	bs.httpserver = &http.Server{Addr: ":8090"}
	//http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := bs.httpserver.ListenAndServe()
		if err.Error() != "http: Server closed" {
			panic(err)
		}
		defer func() {
			log.Println("Webserver exited")
		}()
	}()
}

// Wait for a sigterm event or for user to press control c when running interacticely
func (bs *BasicServer) WaitforInterupt() error {

	ch := make(chan struct{})
	handleSig(func() { close(ch) })
	log.Printf("Waiting for user to press control c or sig terminate\n")
	<-ch
	log.Printf("Terminate signal detected, closing job manager\n")
	return bs.shutdown()
}

func handleSig(cleanupwork cleanupfunc) chan struct{} {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Printf("\nhandleSig Received an interrupt, stopping services...\n")
		if cleanupwork != nil {
			cleanupwork()
		}

		close(cleanupDone)
	}()
	return cleanupDone
}

func (bs *BasicServer) shutdown() error {
	defer bs.ctx.Done()
	defer bs.cancel()
	return bs.httpserver.Shutdown(bs.ctx)
}
