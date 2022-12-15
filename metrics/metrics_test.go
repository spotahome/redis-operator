package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/stretchr/testify/assert"

	"github.com/spotahome/redis-operator/metrics"
)

func TestPrometheusMetrics(t *testing.T) {

	tests := []struct {
		name       string
		addMetrics func(rec metrics.Recorder)
		expMetrics []string
		expCode    int
	}{
		{
			name: "Setting OK should give an OK",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterOK("testns", "test")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting Error should give an Error",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterError("testns", "test")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns"} 0`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting Error after ok should give an Error",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterOK("testns", "test")
				rec.SetClusterError("testns", "test")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns"} 0`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting OK after Error should give an OK",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterError("testns", "test")
				rec.SetClusterOK("testns", "test")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Multiple clusters should appear",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterOK("testns", "test")
				rec.SetClusterOK("testns", "test2")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns"} 1`,
				`my_metrics_controller_cluster_ok{name="test2",namespace="testns"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Same name on different namespaces should appear",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterOK("testns1", "test")
				rec.SetClusterOK("testns2", "test")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns1"} 1`,
				`my_metrics_controller_cluster_ok{name="test",namespace="testns2"} 1`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Deleting a cluster should remove it",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterOK("testns1", "test")
				rec.DeleteCluster("testns1", "test")
			},
			expMetrics: []string{},
			expCode:    http.StatusOK,
		},
		{
			name: "Deleting a cluster should remove only the desired one",
			addMetrics: func(rec metrics.Recorder) {
				rec.SetClusterOK("testns1", "test")
				rec.SetClusterOK("testns2", "test")
				rec.DeleteCluster("testns1", "test")
			},
			expMetrics: []string{
				`my_metrics_controller_cluster_ok{name="test",namespace="testns2"} 1`,
			},
			expCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Create the muxer for testing.
			reg := prometheus.NewRegistry()
			rec := metrics.NewRecorder("my_metrics", reg)

			// Add metrics to prometheus.
			test.addMetrics(rec)

			// Make the request to the metrics.
			h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

			resp := w.Result()
			if assert.Equal(test.expCode, resp.StatusCode) {
				body, _ := io.ReadAll(resp.Body)
				// Check all the metrics are present.
				for _, expMetric := range test.expMetrics {
					assert.Contains(string(body), expMetric)
				}
			}
		})
	}
}
