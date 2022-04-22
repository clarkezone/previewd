// Package testserver contains a dummy server implementation for testing metrics and logging
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

// BasicServer object
type BasicServer struct {
	httpserver *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

// CreateBasicServer Create BasicServer object and return
func CreateBasicServer() *BasicServer {
	bs := BasicServer{}
	return &bs
}

// StartListen Start listening for a connection
func (bs *BasicServer) StartListen(secret string) {
	log.Println("starting... basic server on :8090")

	// prometheus.MustRegister(requestDuration)

	bs.ctx, bs.cancel = context.WithCancel(context.Background())
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		// increment a counter for number of requests processed
	})
	bs.httpserver = &http.Server{Addr: ":8090"}
	// http.Handle("/metrics", promhttp.Handler())
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

// WaitforInterupt Wait for a sigterm event or for user to press control c when running interacticely
func (bs *BasicServer) WaitforInterupt() error {
	ch := make(chan struct{})
	handleSig(func() { close(ch) })
	log.Printf("Waiting for user to press control c or sig terminate\n")
	<-ch
	log.Printf("Terminate signal detected, closing job manager\n")
	return bs.Shutdown()
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

// Shutdown terminates the listening thread
func (bs *BasicServer) Shutdown() error {
	defer bs.ctx.Done()
	defer bs.cancel()
	return bs.httpserver.Shutdown(bs.ctx)
}
