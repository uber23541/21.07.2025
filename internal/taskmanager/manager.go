package taskmanager

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"

	"archive_service/internal/config"
	"archive_service/internal/models"
)

type Manager struct {
	mutex          sync.RWMutex
	tasks          map[string]*models.Task
	activeTasksCnt int
	cfg            *config.Config
	semaphore      chan struct{}
}

// 0755
// owner read, write, execute,
// other read execute.
func New(cfg *config.Config) *Manager {
	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		panic(err)
	}
	return &Manager{
		tasks:     make(map[string]*models.Task),
		cfg:       cfg,
		semaphore: make(chan struct{}, cfg.MaxTasks),
	}
}

func (m *Manager) CreateTask() (*models.Task, error) {
	m.mutex.RLock()
	if m.activeTasksCnt >= m.cfg.MaxTasks {
		m.mutex.RUnlock()
		return nil, fmt.Errorf("server is busy, try later")
	}
	m.mutex.RUnlock()

	task := &models.Task{
		ID:     uuid.New().String(),
		Status: models.StatusPending,
	}

	m.mutex.Lock()
	m.tasks[task.ID] = task
	m.activeTasksCnt++
	m.mutex.Unlock()

	return task, nil
}

func (m *Manager) AddFile(id string, link string) (*models.Task, error) {
	m.mutex.RLock()
	task, ok := m.tasks[id]
	m.mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("task not found")
	}
	if task.Status != models.StatusPending {
		return nil, fmt.Errorf("task already %s", task.Status)
	}

	ext := strings.ToLower(filepath.Ext(link))
	if !contains(m.cfg.AllowedExtensions, ext) {
		return nil, fmt.Errorf("file extension %s is not allowed", ext)
	}

	m.mutex.Lock()
	task.Files = append(task.Files, link)
	isNeedProcess := len(task.Files) >= m.cfg.MaxFilesPerTask
	m.mutex.Unlock()

	if isNeedProcess {
		go m.process(task)
	}
	return task, nil
}

func (m *Manager) Get(id string) (*models.Task, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	t, ok := m.tasks[id]
	return t, ok
}

func (m *Manager) StorageDir() string {
	return m.cfg.StorageDir
}

func (m *Manager) process(task *models.Task) {
	m.semaphore <- struct{}{}
	defer func() { <-m.semaphore }()

	m.setStatus(task, models.StatusProcessing)

	archPath := filepath.Join(m.cfg.StorageDir, task.ID+".zip")
	out, err := os.Create(archPath)
	if err != nil {
		m.finishTask(task, models.StatusError, fmt.Errorf("zip creation: %w", err).Error())
		return
	}
	defer out.Close()

	zipwr := zip.NewWriter(out)
	for _, link := range task.Files {
		if err := m.addFile(zipwr, link); err != nil {
			m.appendErr(task, err.Error())
		}
	}
	if err := zipwr.Close(); err != nil {
		m.appendErr(task, fmt.Errorf("zip close: %w", err).Error())
	}

	resultStatus := models.StatusComplete
	if len(task.Errors) > 0 && len(task.Files) == 0 {
		resultStatus = models.StatusError
	}

	m.mutex.Lock()
	task.ArchPath = "/archives/" + filepath.Base(archPath)
	task.Status = resultStatus
	m.activeTasksCnt--
	m.mutex.Unlock()
}

func (m *Manager) addFile(zipwr *zip.Writer, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}

	w, err := zipwr.Create(filepath.Base(url))
	if err != nil {
		return fmt.Errorf("zip create: %w", err)
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("zip write: %w", err)
	}
	return nil
}

func (m *Manager) setStatus(task *models.Task, status models.TaskStatus) {
	m.mutex.Lock()
	task.Status = status
	m.mutex.Unlock()
}

func (m *Manager) finishTask(task *models.Task, status models.TaskStatus, msg string) {
	m.mutex.Lock()
	task.Status = status
	task.Errors = append(task.Errors, msg)
	m.activeTasksCnt--
	m.mutex.Unlock()
}

func (m *Manager) appendErr(task *models.Task, msg string) {
	m.mutex.Lock()
	task.Errors = append(task.Errors, msg)
	m.mutex.Unlock()
}

func contains(array []string, item string) bool {
	for _, v := range array {
		if v == item {
			return true
		}
	}
	return false
}
