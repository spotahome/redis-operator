package metrics_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spotahome/redis-operator/pkg/metrics"
)

func TestPrometheusMetrics(t *testing.T) {

	tests := []struct {
		name       string
		addMetrics func(pm *metrics.PromMetrics)
		expMetrics []string
		expCode    int
	}{
		{
			name: "Setting number of clusters should get the correct number of clusters",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClustersCreating(6)
				pm.SetClustersRunning(49)
				pm.SetClustersFailed(1)
				pm.SetClustersRunning(53)
			},
			expMetrics: []string{
				`redis_operator_controller_clusters{state="running"} 53`,
				`redis_operator_controller_clusters{state="failed"} 1`,
				`redis_operator_controller_clusters{state="creating"} 6`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Multiple events on the controller should return the correct different event metrics",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.IncAddEventHandled("cluster1")
				pm.IncAddEventHandled("cluster1")
				pm.IncAddEventHandled("cluster1")
				pm.IncDeleteEventHandled("cluster2")
				pm.IncUpdateEventHandled("cluster1")
				pm.IncUpdateEventHandled("cluster3")
				pm.IncUpdateEventHandled("cluster3")
			},
			expMetrics: []string{
				`redis_operator_controller_event_handled_total{cluster="cluster1",kind="add"} 3`,
				`redis_operator_controller_event_handled_total{cluster="cluster1",kind="update"} 1`,
				`redis_operator_controller_event_handled_total{cluster="cluster2",kind="delete"} 1`,
				`redis_operator_controller_event_handled_total{cluster="cluster3",kind="update"} 2`,
			},
			expCode: http.StatusOK,
		},
		{
			name: "Setting number of masters of different clusters should get the correct number of clusters",
			addMetrics: func(pm *metrics.PromMetrics) {
				pm.SetClusterMasters(1, "cluster1")
				pm.SetClusterMasters(0, "cluster2")
				pm.SetClusterMasters(2, "cluster3")
				pm.SetClusterMasters(1, "cluster4")
				pm.SetClusterMasters(3, "cluster1")

			},
			expMetrics: []string{
				`redis_operator_cluster_masters{cluster="cluster1"} 3`,
				`redis_operator_cluster_masters{cluster="cluster2"} 0`,
				`redis_operator_cluster_masters{cluster="cluster3"} 2`,
				`redis_operator_cluster_masters{cluster="cluster4"} 1`,
			},
			expCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			path := "/awesome-metrics"

			// Create the muxer for testing.
			mx := http.NewServeMux()
			pm := metrics.NewPrometheusMetrics(path, mx)

			// Add metrics to prometheus.
			test.addMetrics(pm)

			// Make the request to the metrics.
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			mx.ServeHTTP(w, req)

			resp := w.Result()
			if assert.Equal(test.expCode, resp.StatusCode) {
				body, _ := ioutil.ReadAll(resp.Body)
				// Check all the metrics are present.
				for _, expMetric := range test.expMetrics {
					assert.Contains(string(body), expMetric)
				}
			}
		})
	}
}
