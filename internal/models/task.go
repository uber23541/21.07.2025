package models

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusComplete   TaskStatus = "complete"
	StatusError      TaskStatus = "error"
)

type Task struct {
	ID       string     `json:"id"`
	Files    []string   `json:"files"`
	Status   TaskStatus `json:"status"`
	ArchPath string     `json:"arch_path,omitempty"`
	Errors   []string   `json:"errors,omitempty"`
}
