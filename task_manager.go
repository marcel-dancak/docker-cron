package dcron

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/robfig/cron/v3"
)

type strSlice []string

func (e *strSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value interface{}
	if err := unmarshal(&value); err != nil {
		return err
	}
	switch cmd := value.(type) {
	case string:
		*e = []string{cmd}
		// *e = strings.Fields(cmd)
	case []interface{}:
		array := make([]string, len(cmd))
		for i, item := range cmd {
			array[i] = item.(string)
		}
		*e = array
	default:
		return fmt.Errorf("Unsupported type")
	}
	return nil
}

type baseTask struct {
	Schedule string   `yaml:"schedule"`
	Command  strSlice `yaml:"command"`
}

type runTask struct {
	baseTask    `yaml:",inline"`
	Image       string   `yaml:"image"`
	Volumes     []string `yaml:"volumes,flow"`
	NetworkMode string   `yaml:"network_mode"`
	Entrypoint  strSlice `yaml:"entrypoint"`
}

type execTask struct {
	baseTask `yaml:",inline"`
	Service  string `yaml:"service"`
	User     string `yaml:"user"`
}

// TasksConfig tasks definitions
type TasksConfig struct {
	Run  map[string]runTask
	Exec map[string]execTask
}

// Logger interface for docker tasks
type Logger interface {
	StdoutWriter() io.Writer
	StderrWriter() io.Writer
}

// Task definition
type Task struct {
	Name     string
	Schedule string
	Run      func(l Logger) (int, error)
	EntryID  cron.EntryID
}

type taskListeners struct {
	Started  []func(*Task)
	Finished []func(*Task)
}

// TaskStats stats about run task
type TaskStats struct {
	ID         int       `json:"id"`
	StartTime  time.Time `json:"start_time"`
	Running    bool      `json:"running"`
	Crashed    bool      `json:"crashed"`
	Status     int       `json:"status"`
	StdoutSize int       `json:"stdout_size"`
	StderrSize int       `json:"stderr_size"`
}

type tasksStats struct {
	sync.RWMutex
	Tasks map[string][]*TaskStats
}

// TaskManager export
type TaskManager struct {
	Ctx         context.Context
	Cli         *client.Client
	Cron        *cron.Cron
	ProjectName string
	Tasks       map[string]*Task
	Config      TasksConfig
	Stats       *tasksStats
	LogsRoot    string
	running     bool
	listeners   taskListeners
}

// NewTaskManager export
func NewTaskManager(config TasksConfig, project, logsDir string) (*TaskManager, error) {
	ctx := context.Background()
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		return nil, err
	}
	c := cron.New()
	tm := TaskManager{
		ProjectName: project,
		Ctx:         ctx,
		Cli:         cli,
		Cron:        c,
		LogsRoot:    logsDir,
	}
	tm.listeners = taskListeners{}
	tm.LoadConfig(config)
	return &tm, nil
}

// AddTaskStartedListener register listener for started tasks events
func (m *TaskManager) AddTaskStartedListener(listener func(*Task)) {
	m.listeners.Started = append(m.listeners.Started, listener)
}

// AddTaskFinishedListener register listener for finished tasks events
func (m *TaskManager) AddTaskFinishedListener(listener func(*Task)) {
	m.listeners.Finished = append(m.listeners.Finished, listener)
}

func (m *TaskManager) runTaskFunction(config runTask) func(Logger) (int, error) {
	return func(l Logger) (int, error) {
		return m.runDockerCommand(l, config)
	}
}

func (m *TaskManager) execTaskFunction(config execTask) func(Logger) (int, error) {
	return func(l Logger) (int, error) {
		return m.execDockerCommand(l, config)
	}
}

func (m *TaskManager) cronTask(task *Task) func() {
	return func() {
		m.RunTask(task)
	}
}

// LoadConfig load tasks configuration (without starting)
func (m *TaskManager) LoadConfig(config TasksConfig) {
	tasks := make(map[string]*Task)
	for name, task := range config.Run {
		tasks[name] = &Task{name, task.Schedule, m.runTaskFunction(task), -1}
	}
	for name, task := range config.Exec {
		tasks[name] = &Task{name, task.Schedule, m.execTaskFunction(task), -1}
	}
	m.Cron = cron.New()
	m.Tasks = tasks
	m.Config = config
	// if m.running {
	// 	m.Stop()
	// 	m.Start()
	// }
}

func (m *TaskManager) containerName(name string) string {
	if m.ProjectName != "" {
		return fmt.Sprintf("%s_%s", m.ProjectName, name)
	}
	return name
}

func (m *TaskManager) runDockerCommand(logger Logger, conf runTask) (int, error) {
	config := &container.Config{
		Image:        conf.Image,
		Cmd:          []string(conf.Command),
		AttachStderr: true,
		AttachStdout: true,
	}
	if conf.Entrypoint != nil {
		config.Entrypoint = []string(conf.Entrypoint)
	}

	binds := make([]string, 0)
	for _, item := range conf.Volumes {
		if strings.Contains(item, "/") {
			binds = append(binds, item)
		} else {
			binds = append(binds, m.containerName(item))
		}
	}
	network := conf.NetworkMode
	if network == "" {
		network = fmt.Sprintf("%s_default", m.ProjectName)
	}
	hostConfig := &container.HostConfig{
		Binds:       binds,
		NetworkMode: container.NetworkMode(network),
	}
	resp, err := m.Cli.ContainerCreate(m.Ctx, config, hostConfig, nil, "")
	if err != nil {
		return -1, err
	}
	if err := m.Cli.ContainerStart(m.Ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return -1, err
	}
	status, err := m.Cli.ContainerWait(m.Ctx, resp.ID)
	if err != nil {
		return -1, err
	}
	out, err := m.Cli.ContainerLogs(m.Ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return -1, err
	}
	defer out.Close()

	if _, err := stdcopy.StdCopy(logger.StdoutWriter(), logger.StderrWriter(), out); err != nil {
		log.Printf("Failed to log task output: %s\n", err)
	}
	if err := m.Cli.ContainerRemove(m.Ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		log.Printf("Failed to remove container: %s\n", resp.ID)
	}
	return int(status), nil
}

func (m *TaskManager) getServiceContainers(name string) ([]types.Container, error) {
	list := make([]types.Container, 0)
	query := filters.NewArgs()
	query.Add("status", "running")
	containers, err := m.Cli.ContainerList(m.Ctx, types.ContainerListOptions{Filters: query})
	if err != nil {
		return list, err
	}
	prefix := fmt.Sprintf("/%s", m.containerName(name))
	for _, container := range containers {
		if strings.HasPrefix(container.Names[0], prefix) {
			list = append(list, container)
		}
	}
	return list, nil
}

func (m *TaskManager) execDockerCommand(logger Logger, conf execTask) (int, error) {
	containers, err := m.getServiceContainers(conf.Service)
	if err != nil {
		return -1, err
	}
	for _, container := range containers {
		config := types.ExecConfig{
			User:         conf.User,
			Cmd:          conf.Command,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
			// Detach:       false,
		}
		resp, err := m.Cli.ContainerExecCreate(m.Ctx, container.ID, config)
		if err != nil {
			return -1, err
		}
		atinfo, err := m.Cli.ContainerExecAttach(m.Ctx, resp.ID, config)
		if err != nil {
			return -1, err
		}
		defer atinfo.Close()
		if err := m.Cli.ContainerExecStart(m.Ctx, resp.ID, types.ExecStartCheck{}); err != nil {
			return -1, err
		}
		if _, err := stdcopy.StdCopy(logger.StdoutWriter(), logger.StderrWriter(), atinfo.Reader); err != nil {
			log.Printf("Failed to log task output: %s\n", err)
		}
		inspect, err := m.Cli.ContainerExecInspect(m.Ctx, resp.ID)
		if err != nil {
			return -1, err
		}
		return inspect.ExitCode, nil
	}
	return -1, fmt.Errorf("Running service not found: %s", conf.Service)
}

// GetLogfilePath location of logfile
func (m *TaskManager) GetLogfilePath(task string, id int) string {
	logfilename := fmt.Sprintf("%s.%d.log", task, id)
	return filepath.Join(m.LogsRoot, logfilename)
}

func (m *TaskManager) isTaskRunning(task string) bool {
	m.Stats.Lock()
	defer m.Stats.Unlock()
	if taskStats, ok := m.Stats.Tasks[task]; ok {
		for i := len(taskStats) - 1; i >= 0; i-- {
			if taskStats[i].Running {
				return true
			}
		}
	}
	return false
}

// RunTask execute task
func (m *TaskManager) RunTask(task *Task) {
	startTime := time.Now()
	log.Printf("[CRON] (%s) Start\n", task.Name)
	m.Stats.Lock()
	statsEntry := &TaskStats{
		StartTime: startTime,
		Running:   true,
		Status:    -1,
	}
	if taskStats, ok := m.Stats.Tasks[task.Name]; ok {
		last := taskStats[len(taskStats)-1]
		statsEntry.ID = last.ID + 1
		taskStats = append(taskStats, statsEntry)
		m.Stats.Tasks[task.Name] = taskStats
	} else {
		newStats := make([]*TaskStats, 1)
		statsEntry.ID = 1
		newStats[0] = statsEntry
		m.Stats.Tasks[task.Name] = newStats
	}
	logfile := m.GetLogfilePath(task.Name, statsEntry.ID)
	m.Stats.Unlock()
	f, err := os.Create(logfile)
	if err != nil {
		log.Printf("Failed to create logfile %s: %s\n", logfile, err)
		return
	}
	defer f.Close()
	for _, listener := range m.listeners.Started {
		listener(task)
	}

	logWriter := bufio.NewWriter(f)
	logger := newDockerLogger(logWriter)
	status, err := task.Run(logger)
	m.Stats.Lock()
	if err != nil {
		log.Printf("[CRON] (%s) Error: %s\n", task.Name, err)
		fmt.Fprintf(logger.StderrWriter(), "[CRON] Error: %s\n", err)
		statsEntry.Crashed = true
	} else {
		statsEntry.Status = status
	}
	logWriter.Flush()
	statsEntry.StdoutSize = logger.stdout.Size
	statsEntry.StderrSize = logger.stderr.Size

	statsEntry.Running = false
	m.Stats.Unlock()
	log.Printf("[CRON] (%s) Status: %d\n", task.Name, status)

	go func() {
		time.Sleep(50 * time.Millisecond)
		for _, listener := range m.listeners.Finished {
			listener(task)
		}
	}()
}

// Start start's tasks scheduler
func (m *TaskManager) Start() error {
	if m.Stats == nil {
		stats := &tasksStats{Tasks: make(map[string][]*TaskStats, len(m.Tasks))}
		m.Stats = stats
	}
	for _, task := range m.Tasks {
		if task.Schedule != "" {
			id, err := m.Cron.AddFunc(task.Schedule, m.cronTask(task))
			if err != nil {
				return err
			}
			task.EntryID = id
		}
	}
	m.Cron.Start()
	m.running = true
	return nil
}

// Stop stop's tasks scheduler
func (m *TaskManager) Stop() {
	m.Cron.Stop()
	m.running = false
}
