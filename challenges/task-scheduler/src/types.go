package taskscheduler

import (
	"time"
)

// Priority definiert die Prioritätsstufen für Tasks
type Priority int

const (
	PriorityLow      Priority = 0
	PriorityMedium   Priority = 1
	PriorityHigh     Priority = 2
	PriorityCritical Priority = 3
)

// String gibt die Priorität als lesbaren String zurück
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "LOW"
	case PriorityMedium:
		return "MEDIUM"
	case PriorityHigh:
		return "HIGH"
	case PriorityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// IsValid prüft ob die Priorität gültig ist
func (p Priority) IsValid() bool {
	return p >= PriorityLow && p <= PriorityCritical
}

// TaskStatus repräsentiert den Lebenszyklus-Status eines Tasks
type TaskStatus int

const (
	StatusPending   TaskStatus = 0
	StatusRunning   TaskStatus = 1
	StatusCompleted TaskStatus = 2
	StatusFailed    TaskStatus = 3
	StatusRetrying  TaskStatus = 4
	StatusCancelled TaskStatus = 5
)

// String gibt den Status als lesbaren String zurück
func (s TaskStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusRunning:
		return "RUNNING"
	case StatusCompleted:
		return "COMPLETED"
	case StatusFailed:
		return "FAILED"
	case StatusRetrying:
		return "RETRYING"
	case StatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// IsTerminal gibt true zurück wenn der Status ein Endzustand ist
func (s TaskStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// Task repräsentiert eine planbare Arbeitseinheit
type Task struct {
	ID           string                 // Eindeutige Task-ID
	Name         string                 // Menschenlesbarer Name
	Priority     Priority               // Priorität für die Ausführungsreihenfolge
	Status       TaskStatus             // Aktueller Lebenszyklus-Status
	Handler      string                 // Name des registrierten Handlers
	Payload      map[string]interface{} // Beliebige Daten für den Handler
	Dependencies []string               // Task-IDs die vorher abgeschlossen sein müssen
	MaxRetries   int                    // Maximale Anzahl Wiederholungsversuche
	RetryCount   int                    // Aktuelle Anzahl Wiederholungsversuche
	CreatedAt    time.Time              // Erstellungszeitpunkt
	StartedAt    time.Time              // Startzeitpunkt der Ausführung
	CompletedAt  time.Time              // Abschlusszeitpunkt
	Error        error                  // Letzter Fehler (falls aufgetreten)
	Result       interface{}            // Ergebnis der Ausführung
}

// TaskHandler ist eine Funktion die einen Task verarbeitet
type TaskHandler func(task *Task) error

// RetryPolicy definiert das Retry-Verhalten für fehlgeschlagene Tasks
type RetryPolicy struct {
	MaxRetries int           // Maximale Wiederholungsversuche
	BaseDelay  time.Duration // Basis-Wartezeit zwischen Versuchen
	MaxDelay   time.Duration // Maximale Wartezeit (Cap für Backoff)
	Multiplier float64       // Multiplikator für exponentielles Backoff
}

// SchedulerConfig enthält die Konfiguration für den Scheduler
type SchedulerConfig struct {
	MaxWorkers      int           // Maximale Anzahl gleichzeitiger Worker
	QueueSize       int           // Kapazität der Task-Queue
	DefaultRetry    RetryPolicy   // Standard-Retry-Policy
	ShutdownTimeout time.Duration // Timeout für graceful Shutdown
}

// TaskFilter ermöglicht das Filtern von Tasks im Store
type TaskFilter struct {
	Status   *TaskStatus // Filter nach Status (nil = alle)
	Priority *Priority   // Filter nach Priorität (nil = alle)
	Handler  string      // Filter nach Handler-Name (leer = alle)
}

// TaskMetrics enthält Statistiken über Task-Ausführungen
type TaskMetrics struct {
	TotalSubmitted  int64                // Gesamtanzahl eingereichter Tasks
	TotalCompleted  int64                // Gesamtanzahl abgeschlossener Tasks
	TotalFailed     int64                // Gesamtanzahl fehlgeschlagener Tasks
	TotalRetried    int64                // Gesamtanzahl wiederholter Tasks
	AverageExecTime time.Duration        // Durchschnittliche Ausführungszeit
	TasksByPriority map[Priority]int64   // Anzahl Tasks pro Priorität
	TasksByStatus   map[TaskStatus]int64 // Anzahl Tasks pro Status
}

// RetryableError signalisiert einen Fehler der einen Retry auslösen soll
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NonRetryableError signalisiert einen Fehler der KEINEN Retry auslösen soll
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}
