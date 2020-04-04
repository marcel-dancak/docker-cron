package dcron

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
)

// Server web server API for task manager
type Server struct {
	router      *chi.Mux
	taskManager *TaskManager
}

// PublicServer web server with websocket connections
type PublicServer struct {
	Server
	upgrader websocket.Upgrader
	hub      *Hub
}

type taskInfo struct {
	Name     string      `json:"name"`
	Schedule string      `json:"schedule,omitempty"`
	Next     *time.Time  `json:"next,omitempty"`
	Stats    []TaskStats `json:"stats"`
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Failed to serialize JSON: %s\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
	}
}

func (s *Server) getTaskInfo(task *Task) taskInfo {
	s.taskManager.Stats.RLock()
	stats := s.taskManager.Stats.Tasks[task.Name]
	copy := make([]TaskStats, len(stats))
	for i, value := range stats {
		copy[i] = *value
	}
	s.taskManager.Stats.RUnlock()

	var next *time.Time
	if task.Schedule != "" {
		entry := s.taskManager.Cron.Entry(task.EntryID)
		next = &entry.Next
	}
	return taskInfo{
		task.Name,
		task.Schedule,
		next,
		copy,
	}
}

// func (s *Server) handleTasksConfig(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/yaml")
// 	yaml.NewEncoder(w).Encode(s.taskManager.Config)
// }

func (s *Server) handleTasksInfo(w http.ResponseWriter, r *http.Request) {
	tasks := make(map[string]taskInfo, len(s.taskManager.Tasks))
	for name, task := range s.taskManager.Tasks {
		tasks[name] = s.getTaskInfo(task)
	}
	s.jsonResponse(w, tasks)
}

func (s *Server) handleTasksList(w http.ResponseWriter, r *http.Request) {
	tasks := make([]taskInfo, 0)
	for _, task := range s.taskManager.Tasks {
		tasks = append(tasks, s.getTaskInfo(task))
	}
	s.jsonResponse(w, tasks)
}

func (s *Server) handleTaskRun(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "task")
	task, ok := s.taskManager.Tasks[name]
	if !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if s.taskManager.isTaskRunning(name) {
		http.Error(w, "Task is already running", http.StatusConflict)
		return
	}
	go s.taskManager.RunTask(task)
	fmt.Fprintf(w, "ok\n")
}

func (s *Server) handleTaskLogs(w http.ResponseWriter, r *http.Request) {
	task := chi.URLParam(r, "task")
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	logfile := s.taskManager.GetLogfilePath(task, id)
	http.ServeFile(w, r, logfile)
}

func (s *Server) handleKillService(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")
	signal := r.URL.Query().Get("signal")
	tm := s.taskManager
	containers, err := tm.getServiceContainers(service)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if len(containers) == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}
	for _, container := range containers {
		if err := tm.Cli.ContainerKill(tm.Ctx, container.ID, signal); err != nil {
			log.Printf(`Failed to send signal to service "%s": %s\n`, service, err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintf(w, "ok\n")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// trailing slashes policy
	// if r.URL.Path == "/ui" {
	// 	r.URL.Path = "/ui/"
	// }
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
	}
	s.router.ServeHTTP(w, r)
}

func (s *PublicServer) handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "WS connection failed", http.StatusInternalServerError)
		return
	}
	client := &Client{hub: s.hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
}

func (s *PublicServer) broadcastJSON(data interface{}) {
	json, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	s.hub.broadcast <- json
}

type taskNotificationMessage struct {
	Type string   `json:"type"`
	Task taskInfo `json:"task"`
}

func (s *PublicServer) taskStarted(task *Task) {
	taskInfo := s.getTaskInfo(task)
	msg := taskNotificationMessage{"TaskStarted", taskInfo}
	s.broadcastJSON(msg)
}

func (s *PublicServer) taskFinished(task *Task) {
	taskInfo := s.getTaskInfo(task)
	msg := taskNotificationMessage{"TaskFinished", taskInfo}
	s.broadcastJSON(msg)
}

func indexHandler(webRoot string) func(w http.ResponseWriter, r *http.Request) {
	indexHTML := filepath.Join(webRoot, "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexHTML)
	}
}

// NewServer creates a new API server
func NewServer(taskManager *TaskManager) *Server {
	router := chi.NewRouter()
	s := Server{router, taskManager}
	api := router.Group(nil)
	api.Use(middleware.Logger)
	api.HandleFunc("/api/tasks", s.handleTasksInfo)
	api.Post("/api/run/{task}", s.handleTaskRun)
	api.Post("/api/services/kill/{service}", s.handleKillService)
	api.HandleFunc("/api/logs/{task}/{id:[0-9]+}", s.handleTaskLogs)
	return &s
}

// NewPublicServer creates a new public server
func NewPublicServer(taskManager *TaskManager, webRoot string, auth AuthBackend) *PublicServer {
	router := chi.NewRouter()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	wsHub := newHub()
	go wsHub.run()
	s := PublicServer{Server{router, taskManager}, upgrader, wsHub}

	api := router.Group(nil)
	api.Use(middleware.Logger)

	if auth != nil {
		router.Post("/api/auth/login", auth.HandleLogin)
		api.Use(auth.AuthMiddleware)
	}

	api.HandleFunc("/api/tasks", s.handleTasksInfo)
	api.Post("/api/run/{task}", s.handleTaskRun)
	api.HandleFunc("/api/logs/{task}/{id:[0-9]+}", s.handleTaskLogs)
	api.HandleFunc("/ws", s.handleWs)
	router.Handle("/ui/static/*", http.StripPrefix("/ui/", http.FileServer(http.Dir(webRoot))))
	router.HandleFunc("/ui", indexHandler(webRoot))
	// for SPA with router in history mode
	router.HandleFunc("/ui/*", indexHandler(webRoot))
	// indexHTML := filepath.Join(webRoot, "index.html")
	// router.HandleFunc("/ui/*", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, indexHTML)
	// })

	s.taskManager.AddTaskStartedListener(s.taskStarted)
	s.taskManager.AddTaskFinishedListener(s.taskFinished)
	return &s
}
