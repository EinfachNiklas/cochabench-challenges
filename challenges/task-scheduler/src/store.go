package taskscheduler

import (
	"errors"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrTaskIDEmpty       = errors.New("task ID must not be empty")
)

type TaskStore struct {
}

func NewTaskStore() *TaskStore {
	return nil
}

func (s *TaskStore) Add(task *Task) error {
	return nil
}

func (s *TaskStore) Get(id string) (*Task, error) {
	return nil, ErrTaskNotFound
}

func (s *TaskStore) Update(task *Task) error {
	return ErrTaskNotFound
}

func (s *TaskStore) Delete(id string) error {
	return ErrTaskNotFound
}

func (s *TaskStore) List(filter TaskFilter) []*Task {
	return nil
}

func (s *TaskStore) Count() int {
	return 0
}
