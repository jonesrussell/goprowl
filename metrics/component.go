package metrics

import (
	"fmt"
	"time"
)

// ComponentMetrics holds metrics for any application component
type ComponentMetrics struct {
	collector   *MetricsCollector
	componentID string
}

// NewComponentMetrics creates metrics for a specific component
func NewComponentMetrics(collector *MetricsCollector, componentType string) *ComponentMetrics {
	id := fmt.Sprintf("%s-%d", componentType, time.Now().UnixNano())
	return &ComponentMetrics{
		collector:   collector,
		componentID: id,
	}
}

// Crawler-specific methods
func (m *ComponentMetrics) IncrementActiveRequests() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.totalActiveRequests.WithLabelValues(m.componentID).Inc()
}

func (m *ComponentMetrics) DecrementActiveRequests() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.totalActiveRequests.WithLabelValues(m.componentID).Dec()
}

func (m *ComponentMetrics) IncrementErrorCount() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.totalErrors.WithLabelValues(m.componentID).Inc()
}

func (m *ComponentMetrics) IncrementPagesProcessed() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.totalPagesProcessed.WithLabelValues(m.componentID).Inc()
}

func (m *ComponentMetrics) ObserveResponseSize(size float64) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.responseSizes.WithLabelValues(m.componentID).Observe(size)
}

func (m *ComponentMetrics) ObserveRequestDuration(duration float64) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.requestDurations.WithLabelValues(m.componentID).Observe(duration)
}

// ... other methods ...
