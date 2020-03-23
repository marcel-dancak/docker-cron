package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/marcel-dancak/dcron"
	"gopkg.in/yaml.v2"
)

func optEnv(key, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defaultValue
}

func parseConfig(path string) (dcron.TasksConfig, error) {
	config := dcron.TasksConfig{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	return config, nil
}

func main() {
	configPath := flag.String("f", os.Getenv("DCRON_CONFIG_FILE"), "tasks configuration file")
	// projectName := flag.String("project", os.Getenv("DCRON_COMPOSE_PROJECT"), "Docker Compose project's name")
	flag.Parse()
	if *configPath == "" {
		log.Fatal("Config file not specified!")
	}

	config, err := parseConfig(*configPath)
	if err != nil {
		log.Printf("[CRON] Failed to parse config file: %s\n", *configPath)
		log.Fatal(err)
	}

	projectName := os.Getenv("DCRON_COMPOSE_PROJECT")
	if projectName == "" {
		log.Fatal("Docker Compose project's name not specified!")
	}
	logsDir := filepath.Join(optEnv("DCRON_LOGS_ROOT", "/var/log/dcron"), projectName)
	tm, err := dcron.NewTaskManager(config, projectName, logsDir)
	if err != nil {
		log.Println("Failed to initialize Task Manager")
		log.Fatal(err)
	}

	tm.Start()
	log.Println("[CRON] Starting Cron Jobs")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Reloading configuration file:", event.Name)
					conf, err := parseConfig(*configPath)
					if err != nil {
						log.Printf("[CRON] Failed to parse config file: %s\n", *configPath)
						log.Println(err)
					} else {
						tm.Stop()
						tm.LoadConfig(conf)
						tm.Start()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	watcher.Add(*configPath)

	address := fmt.Sprintf(":%s", optEnv("DCRON_WEB_PORT", "8090"))
	server := dcron.NewServer(tm, optEnv("DCRON_WEB_ROOT", "/var/www"))
	// baseCtx := valv.Context()
	// srv := http.Server{
	// 	Addr: address,
	// 	Handler: chi.ServerBaseContext(baseCtx, r),
	// }
	log.Fatal(http.ListenAndServe(address, server))

	// done := make(chan os.Signal, 1)
	// signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	// <-done

	tm.Stop()
}
