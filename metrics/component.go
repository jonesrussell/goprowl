package metrics

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ComponentMetrics holds metrics for any application component
type ComponentMetrics struct {
	collector     *MetricsCollector
	componentID   string
	componentType string
}

// NewComponentMetrics creates metrics for a specific component
func NewComponentMetrics(collector *MetricsCollector, componentType string) *ComponentMetrics {
	id := fmt.Sprintf("component-%d", time.Now().UnixNano())
	return &ComponentMetrics{
		collector:     collector,
		componentID:   id,
		componentType: componentType,
	}
}

// IncrementActiveRequests increments the active requests metric
func (m *ComponentMetrics) IncrementActiveRequests() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("incrementing active requests",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalActiveRequests.WithLabelValues(m.componentID, m.componentType).Inc()
}

// DecrementActiveRequests decrements the active requests metric
func (m *ComponentMetrics) DecrementActiveRequests() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("decrementing active requests",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalActiveRequests.WithLabelValues(m.componentID, m.componentType).Dec()
}

// IncrementErrorCount increments the error count metric
func (m *ComponentMetrics) IncrementErrorCount() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("incrementing error count",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalErrors.WithLabelValues(m.componentID, m.componentType).Inc()
}

// IncrementPagesProcessed increments the pages processed metric
func (m *ComponentMetrics) IncrementPagesProcessed() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("incrementing pages processed",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalPagesProcessed.WithLabelValues(m.componentID, m.componentType).Inc()
}

// ObserveResponseSize observes the response size
func (m *ComponentMetrics) ObserveResponseSize(size float64) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("observing response size",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.responseSizes.WithLabelValues(m.componentID).Observe(size)
}

// ObserveRequestDuration observes the request duration
func (m *ComponentMetrics) ObserveRequestDuration(duration float64) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("observing request duration",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.requestDurations.WithLabelValues(m.componentID, m.componentType).Observe(duration)
}

// ComponentMetrics methods for list operations

// ObserveHistogram observes the specified histogram metric
func (m *ComponentMetrics) ObserveHistogram(name string, value float64) {
	m.collector.mu.Lock() // Use the collector's mutex instead of undefined lockCollector
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("observing histogram",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))

	switch name {
	case "list_documents_duration_seconds":
		m.collector.listOperationDuration.WithLabelValues(m.componentID).Observe(value)
	default:
		m.collector.logger.Warn("unknown histogram metric", zap.String("name", name))
	}
}

// IncCounter increments the specified counter metric
func (m *ComponentMetrics) IncCounter(name string, value float64) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("incrementing counter",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))

	switch name {
	case "list_documents_errors_total":
		m.collector.listOperationErrors.WithLabelValues(m.componentID).Add(value)
	default:
		m.collector.logger.Warn("unknown counter metric", zap.String("name", name))
	}
}

// SetGaugeValue sets the specified gauge metric
func (m *ComponentMetrics) SetGaugeValue(name string, value float64) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("setting gauge value",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))

	switch name {
	case "indexed_documents_total":
		m.collector.indexedDocuments.WithLabelValues(m.componentID).Set(value)
	default:
		m.collector.logger.Warn("unknown gauge metric", zap.String("name", name))
	}
}

// Add the new methods required by the crawler

// IncrementActiveRequestsWithLabel increments the active requests metric with a label
func (m *ComponentMetrics) IncrementActiveRequestsWithLabel(label string) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("incrementing active requests with label",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalActiveRequests.WithLabelValues(label, m.componentType).Inc()
}

// IncrementPagesProcessedWithLabel increments the pages processed metric with a label
func (m *ComponentMetrics) IncrementPagesProcessedWithLabel(label string) {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("incrementing pages processed with label",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalPagesProcessed.WithLabelValues(label, m.componentType).Inc()
}

// ResetActiveRequests resets the active requests metric
func (m *ComponentMetrics) ResetActiveRequests() {
	m.collector.mu.Lock()
	defer m.collector.mu.Unlock()
	m.collector.logger.Debug("resetting active requests",
		zap.String("component_id", m.componentID),
		zap.String("component_type", m.componentType))
	m.collector.totalActiveRequests.Reset()
}
