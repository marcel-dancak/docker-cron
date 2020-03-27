package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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
	configPath := os.Getenv("DCRON_CONFIG_FILE")
	projectName := os.Getenv("DCRON_COMPOSE_PROJECT")
	if configPath == "" {
		log.Fatal("Config file not specified!")
	}
	if projectName == "" {
		log.Fatal("Docker Compose project's name not specified!")
	}

	config, err := parseConfig(configPath)
	if err != nil {
		log.Printf("[CRON] Failed to parse config file: %s\n", configPath)
		log.Fatal(err)
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
		pendingReload := false
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if !pendingReload {
						pendingReload = true
						time.AfterFunc(1*time.Second, func() {
							log.Println("Reloading configuration file:", event.Name)
							conf, err := parseConfig(configPath)
							if err != nil {
								log.Printf("[CRON] Failed to parse config file: %s\n", configPath)
								log.Println(err)
							} else {
								tm.Stop()
								tm.LoadConfig(conf)
								tm.Start()
							}
							pendingReload = false
						})
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
	watcher.Add(configPath)

	webPort, webServerConfigured := os.LookupEnv("DCRON_WEB_PORT")
	if webServerConfigured {
		webAddress := fmt.Sprintf(":%s", webPort)
		publicServer := dcron.NewPublicServer(tm, optEnv("DCRON_WEB_ROOT", "/var/www"))
		certFile, useSSL := os.LookupEnv("DCRON_SSL_CERT")
		log.Println("[CRON] Starting public web server on port:", webPort)
		if useSSL {
			keyFile, _ := os.LookupEnv("DCRON_SSL_CERT_KEY")
			go http.ListenAndServeTLS(webAddress, certFile, keyFile, publicServer)
		} else {
			go http.ListenAndServe(webAddress, publicServer)
		}
	}

	apiServer := dcron.NewServer(tm)
	apiPort := optEnv("DCRON_API_PORT", "7000")
	log.Println("[CRON] Starting API server on port:", apiPort)
	go http.ListenAndServe(fmt.Sprintf(":%s", apiPort), apiServer)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done

	tm.Stop()
}
