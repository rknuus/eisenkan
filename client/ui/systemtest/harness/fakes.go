package harness

import "time"

// FakeWorkflowManager provides minimal behaviors for tests.
type FakeWorkflowManager struct {
	Delay             time.Duration
	ShouldFail        bool
	ValidationMessage string
}

func (f *FakeWorkflowManager) MoveTask(taskID, from, to string) error {
	time.Sleep(f.Delay)
	if f.ShouldFail {
		return &FakeError{Message: f.ValidationMessage}
	}
	return nil
}

func (f *FakeWorkflowManager) CreateTask(data any) (string, error) {
	time.Sleep(f.Delay)
	if f.ShouldFail {
		return "", &FakeError{Message: f.ValidationMessage}
	}
	return "task-id-1", nil
}

// FakeFormValidationEngine simulates validation outcomes.
type FakeFormValidationEngine struct {
	InvalidFields map[string]string
}

func (f *FakeFormValidationEngine) Validate(fields map[string]string) map[string]string {
	// Merge configured invalid fields onto input as errors.
	out := map[string]string{}
	for k, v := range f.InvalidFields {
		out[k] = v
	}
	return out
}

// FakeDragDropEngine provides noop spatial mechanics for tests.
type FakeDragDropEngine struct{}

func (f *FakeDragDropEngine) Allow(from, to string) bool { return true }

// FakeLayoutEngine placeholder.
type FakeLayoutEngine struct{}

// FakeError is a simple error type used by fakes.
type FakeError struct{ Message string }

func (e *FakeError) Error() string { return e.Message }
