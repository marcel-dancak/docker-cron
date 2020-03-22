package dcron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
)

// Server web server API and interface for task manager
type Server struct {
	router      *chi.Mux
	taskManager *TaskManager
	upgrader    websocket.Upgrader
	wsConns     map[*websocket.Conn]bool
	hub         *Hub
}

type taskInfo struct {
	Name     string      `json:"name"`
	Schedule string      `json:"schedule"`
	Next     time.Time   `json:"next"`
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

func (s *Server) handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "WS connection failed", http.StatusInternalServerError)
		return
	}
	s.wsConns[conn] = true
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		if bytes.Compare(msg, []byte("Ping")) == 0 {
			continue
		}
		log.Println(msgType, string(msg))
	}
	delete(s.wsConns, conn)
}

func (s *Server) handleWs2(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) getTaskInfo(task *Task) taskInfo {
	entry := s.taskManager.Cron.Entry(task.EntryID)
	s.taskManager.Stats.RLock()
	stats := s.taskManager.Stats.Tasks[task.Name]
	copy := make([]TaskStats, len(stats))
	for i, value := range stats {
		copy[i] = *value
	}
	s.taskManager.Stats.RUnlock()
	return taskInfo{
		task.Name,
		task.Schedule,
		entry.Next,
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// func (s *Server) broadcastJSON(data interface{}) {
// 	for conn := range s.wsConns {
// 		if err := conn.WriteJSON(data); err != nil {
// 			log.Printf("Failed to send JSON message: %s\n", err)
// 		}
// 	}
// }

func (s *Server) broadcastJSON(data interface{}) {
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

func (s *Server) taskStarted(task *Task) {
	taskInfo := s.getTaskInfo(task)
	msg := taskNotificationMessage{"TaskStarted", taskInfo}
	s.broadcastJSON(msg)
}

func (s *Server) taskFinished(task *Task) {
	taskInfo := s.getTaskInfo(task)
	msg := taskNotificationMessage{"TaskFinished", taskInfo}
	s.broadcastJSON(msg)
}

// NewServer creates a new server
func NewServer(taskManager *TaskManager, webRoot string) *Server {
	router := chi.NewRouter()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	wsConns := make(map[*websocket.Conn]bool)
	wsHub := newHub()
	go wsHub.run()
	s := Server{router, taskManager, upgrader, wsConns, wsHub}

	api := router.Group(nil)
	api.Use(middleware.Logger)
	api.HandleFunc("/api/tasks/", s.handleTasksInfo)
	api.Post("/api/run/{task}", s.handleTaskRun)
	api.HandleFunc("/api/logs/{task}/{id:[0-9]+}", s.handleTaskLogs)
	api.HandleFunc("/ws/", s.handleWs2)
	router.Handle("/ui/*", http.StripPrefix("/ui/", http.FileServer(http.Dir(webRoot))))

	s.taskManager.AddTaskStartedListener(s.taskStarted)
	s.taskManager.AddTaskFinishedListener(s.taskFinished)
	return &s
}
