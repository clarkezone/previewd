package testserver

import (
	"testing"
)

func Test_webhooklistening(t *testing.T) {
	// wait := make(chan bool)
	wh := BasicServer{}
	wh.StartListen("ss")
	// TODO: how to wait for server async start
	// client := &http.Client{}
	// req, err := http.NewRequestWithContext(context.Background(), "POST", "http://0.0.0.0:8090", nil)
	// if err != nil {
	//	t.Errorf("bad request")
	// }
	// req.Header.Set("X-GitHub-Event", "push")
	// resp, err := client.Do(req)
	// if err != nil {
	//	log.Fatal(err)
	// }
	// <-wait
	err := wh.Shutdown()
	if err != nil {
		t.Errorf("shutdown failed")
	}
	// err = resp.Body.Close()
	// if err != nil {
	//	t.Errorf("close failed")
	//}
}
