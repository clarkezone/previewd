// package testserver contains a dummy server implementation for testing metrics and logging
package testserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

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

type BasicServer struct {
	httpserver *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

func CreateBasicServer() *BasicServer {
	bs := BasicServer{}
	return &bs
}

func (bs *BasicServer) StartListen(secret string) {
	fmt.Println("starting...")

	//prometheus.MustRegister(requestDuration)

	bs.ctx, bs.cancel = context.WithCancel(context.Background())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//start a timer
		//start := time.Now()

		//Call webhooklistener

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

func (bs *BasicServer) Shutdown() error {
	defer bs.ctx.Done()
	defer bs.cancel()
	return bs.httpserver.Shutdown(bs.ctx)
}
