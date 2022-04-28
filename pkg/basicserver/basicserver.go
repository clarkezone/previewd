// Package basicserver contains a dummy server implementation for testing metrics and logging
package basicserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/clarkezone/previewd/internal"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

type cleanupfunc func()

// BasicServer object
type BasicServer struct {
	httpserver *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
	exitchan   chan (bool)
}

// CreateBasicServer Create BasicServer object and return
func CreateBasicServer() *BasicServer {
	bs := BasicServer{}
	return &bs
}

// StartListen Start listening for a connection
func (bs *BasicServer) StartListen(secret string) {
	clarkezoneLog.Successf("starting... basic server on :%v", fmt.Sprint(internal.Port))

	bs.exitchan = make(chan bool)
	// prometheus.MustRegister(requestDuration)
	bs.ctx, bs.cancel = context.WithCancel(context.Background())
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		// increment a counter for number of requests processed
	})

	bs.httpserver = &http.Server{Addr: ":" + fmt.Sprint(internal.Port)}
	// http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := bs.httpserver.ListenAndServe()
		if err.Error() != "http: Server closed" {
			panic(err)
		}
		defer func() {
			clarkezoneLog.Debugf("Webserver goroutine exited")
			bs.exitchan <- true
		}()
	}()
}

// WaitforInterupt Wait for a sigterm event or for user to press control c when running interacticely
func (bs *BasicServer) WaitforInterupt() error {
	if bs.exitchan == nil {
		clarkezoneLog.Debugf("WaitForInterupt(): server not started\n")
		return fmt.Errorf("server not started")
	}
	ch := make(chan struct{})
	handleSig(func() { close(ch) })
	clarkezoneLog.Successf("Waiting for user to press control c or sig terminate\n")
	<-ch
	clarkezoneLog.Debugf("Terminate signal detected, closing job manager\n")
	return bs.Shutdown()
}

func handleSig(cleanupwork cleanupfunc) chan struct{} {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		clarkezoneLog.Debugf("\nhandleSig Received an interrupt, stopping services...\n")
		if cleanupwork != nil {
			clarkezoneLog.Debugf("")
			cleanupwork()
		}

		close(cleanupDone)
	}()
	return cleanupDone
}

// Shutdown terminates the listening thread
func (bs *BasicServer) Shutdown() error {
	if bs.exitchan == nil {
		clarkezoneLog.Debugf("\nno exit channel detected on shutdown\n")
		return fmt.Errorf("no exit channel detected on shutdown")
	}
	defer bs.ctx.Done()
	defer bs.cancel()
	httpexit := bs.httpserver.Shutdown(bs.ctx)
	<-bs.exitchan
	return httpexit
}
