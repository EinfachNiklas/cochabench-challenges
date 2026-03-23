package taskscheduler

import (
	"time"
)

type MetricsCollector struct {
}

func NewMetricsCollector() *MetricsCollector {
	return nil
}

func (mc *MetricsCollector) RecordSubmission(priority Priority) {
}

func (mc *MetricsCollector) RecordCompletion(priority Priority, duration time.Duration) {
}

func (mc *MetricsCollector) RecordFailure(priority Priority) {
}

func (mc *MetricsCollector) RecordRetry(priority Priority) {
}

func (mc *MetricsCollector) RecordStart(priority Priority) {
}

func (mc *MetricsCollector) Snapshot() TaskMetrics {
	return TaskMetrics{
		TasksByPriority: make(map[Priority]int64),
		TasksByStatus:   make(map[TaskStatus]int64),
	}
}
