package metrics

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// MetricsServer holds the configuration and dependencies for the metrics server
type MetricsServer struct {
	logger   *zap.Logger
	config   Config
	registry *prometheus.Registry
}

// NewMetricsServer creates and starts a new MetricsServer
func NewMetricsServer(lc fx.Lifecycle, logger *zap.Logger, config Config, registry *prometheus.Registry) *MetricsServer {
	server := &MetricsServer{
		logger:   logger,
		config:   config,
		registry: registry,
	}

	// Create a new mux for routing
	mux := http.NewServeMux()

	// Register metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Register API endpoints
	mux.HandleFunc("/api/v1/query", server.handleQuery)

	// Register dashboard
	RegisterDashboard(mux)

	// Start HTTP server with the combined mux
	go func() {
		addr := config.MetricsPort
		logger.Info("starting metrics server",
			zap.String("addr", addr),
			zap.String("metrics_url", "http://localhost"+addr+"/metrics"),
			zap.String("api_url", "http://localhost"+addr+"/api/v1/query"),
			zap.String("dashboard_url", "http://localhost"+addr+"/dashboard"),
		)
		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			logger.Fatal("metrics server failed", zap.Error(err))
		}
	}()

	return server
}

// handleQuery handles the /api/v1/query endpoint for querying metrics
func (s *MetricsServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received query request", zap.String("url", r.URL.String()))

	query := r.URL.Query().Get("query")
	if query == "" {
		s.logger.Warn("Query parameter is missing")
		http.Error(w, "query parameter is required", http.StatusBadRequest)
		return
	}

	// Gather metrics
	metrics, err := s.registry.Gather()
	if err != nil {
		s.logger.Error("Failed to gather metrics", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Find matching metric
	var result []struct {
		Metric map[string]string `json:"metric"`
		Value  []interface{}     `json:"value"`
	}

	for _, mf := range metrics {
		if *mf.Name == query {
			for _, m := range mf.Metric {
				labels := make(map[string]string)
				for _, l := range m.Label {
					labels[*l.Name] = *l.Value
				}

				result = append(result, struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				}{
					Metric: labels,
					Value: []interface{}{
						float64(time.Now().Unix()),
						model.SampleValue(*m.Gauge.Value).String(),
					},
				})
			}
		}
	}

	// Prepare response
	response := QueryResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		}{
			ResultType: "vector",
			Result:     result,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode response", zap.Error(err))
	}
}
