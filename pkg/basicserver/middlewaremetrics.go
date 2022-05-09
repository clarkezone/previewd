package basicserver

import (
	"net/http"
	"time"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PromMetricsMiddleware adds simple prometheus metrics type PromMetricsMiddleware
type PromMetricsMiddleware struct {
	handler         http.Handler
	opsProcessed    prometheus.Counter
	requestDuration *prometheus.HistogramVec
}

func (l *PromMetricsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.handler.ServeHTTP(w, r)
	httpDuration := time.Since(start)
	l.opsProcessed.Inc()
	l.requestDuration.WithLabelValues("endpoint", r.URL.RawPath).Observe(httpDuration.Seconds())
}

func newMiddleware(handlerToWrap http.Handler) *PromMetricsMiddleware {
	mw := PromMetricsMiddleware{}
	mw.handler = handlerToWrap
	mw.opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "previewd_testserver_totalops",
		Help: "The total number of processed http requests for testserver",
	})
	mw.requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "previewd_testserver_duration_seconds",
		Help:    "Histogram of duration in seconds",
		Buckets: []float64{1, 2, 5, 7, 10},
	},
		[]string{"endpoint"})
	prometheus.MustRegister(mw.requestDuration)
	return &mw
}

// NewPromMetricsMiddleware constructs a new Logger middleware handler
func NewPromMetricsMiddleware(handlerToWrap http.Handler) *PromMetricsMiddleware {
	clarkezoneLog.Debugf("NewPromMetricsMiddleware()")
	return newMiddleware(handlerToWrap)
}
