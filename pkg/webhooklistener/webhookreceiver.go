// Package webhooklistener provides a webhook that trigger git and job actions
package webhooklistener

import (
	"net/http"

	"github.com/clarkezone/previewd/pkg/basicserver"
	lrm "github.com/clarkezone/previewd/pkg/localrepomanager"

	"github.com/clarkezone/hookserve/hookserve"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

// WebhookListener struct holds state for webhook
type WebhookListener struct {
	lrm          *lrm.LocalRepoManager
	initialBuild bool
	hookserver   *hookserve.Server
	basicServer  *basicserver.BasicServer
	exitchan     chan bool
}

// CreateWebhookListener creates a new instance of WebhookListener
func CreateWebhookListener(lrm *lrm.LocalRepoManager) *WebhookListener {
	wl := WebhookListener{}
	wl.lrm = lrm
	wl.basicServer = basicserver.CreateBasicServer()
	wl.exitchan = make(chan bool)
	return &wl
}

// StartListen creates httpserver to listen for webhook
func (wl *WebhookListener) StartListen(secret string) {
	clarkezoneLog.Infof("Started webhook")

	wl.hookserver = hookserve.NewServer()
	mux := basicserver.DefaultMux()
	mux.HandleFunc("/", wl.getHandler())
	var wrappedMux http.Handler
	wrappedMux = basicserver.NewLoggingMiddleware(mux)
	wrappedMux = basicserver.NewPromMetricsMiddleware("previewd_webhook", wrappedMux)

	go wl.getHookProcessor()()
	wl.basicServer.StartListen(secret, wrappedMux)
}

func (wl *WebhookListener) getHandler() func(w http.ResponseWriter, r *http.Request) {
	responsewriter := func(w http.ResponseWriter, r *http.Request) {
		wl.hookserver.ServeHTTP(w, r)
	}
	return responsewriter
}

func (wl *WebhookListener) getHookProcessor() func() {
	return func() {
		clarkezoneLog.Debugf("WebhookListener: processing loop started")
		defer func() {
			clarkezoneLog.Debugf("WebhookListener: processing loop exited")
		}()
		for {
			select {
			case <-wl.exitchan:
				return
			case event := <-wl.hookserver.Events:
				clarkezoneLog.Debugf(event.Owner + " " + event.Repo + " " + event.Branch + " " + event.Commit)
				if wl.lrm == nil {
					clarkezoneLog.Debugf("WebhookListener: Webhook event ignored as lrm is not initialized")
					break
				}
				err := wl.lrm.HandleWebhook(event.Branch, wl.initialBuild, wl.initialBuild)
				if err != nil {
					clarkezoneLog.Errorf("WebhookListener:HandleWebhook failed:%v", err)
				}
			}
		}
	}
}

// Shutdown closes underlying basicServer
func (wl *WebhookListener) Shutdown() error {
	clarkezoneLog.Debugf("WebHookListener:Shutdown() Shutdown goroutine")
	close(wl.exitchan)
	clarkezoneLog.Debugf("WebHookListener:Shutdown() executed send to exitchan; call shutdown on basicServer")
	err := wl.basicServer.Shutdown()
	clarkezoneLog.Debugf("WebHookListener:Shutdown() shutdown on web returned")
	return err
}
