// Scene Scheduler - OBS Automation Software
// Copyright (c) 2025 Scene Scheduler, S.L. - All Rights Reserved
// This is proprietary software. Unauthorized copying or distribution is prohibited.
// See LICENSE file for terms.

// main.go
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scenescheduler/backend/config"
	"scenescheduler/backend/eventbus"
	"scenescheduler/backend/gui"
	"scenescheduler/backend/logger"
	"scenescheduler/backend/mediasource"
	"scenescheduler/backend/obsclient"
	"scenescheduler/backend/scheduler"
	"scenescheduler/backend/webserver"
)

//go:embed all:frontend/public
var embeddedFS embed.FS

func main() {
	//runtime.GOMAXPROCS(8)

	//*************** 0. Flag Parsing **************************************************
	// Define flags for runtime actions, like listing devices or specifying a config path.
	listDevicesFlag := flag.Bool("list-devices", false, "List all available media devices and exit")
	flag.Parse()

	if *listDevicesFlag {
		mediasource.ShowAllDevices()
		return
	}

	//*************** 1. Setup Context and Graceful Shutdown ***************************
	// The main context controls the lifecycle of all background services.
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		fmt.Fprintf(os.Stderr, "[%s] INFO [main]: Signal %v received. Shutting down...\n", time.Now().Format("15:04:05"), sig)
		cancel() // Signal all goroutines to stop.
	}()

	//*************** 2. Load Global Configuration *************************************
	// The config.New function handles loading, defaults, and validation.
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] CRITICAL [main]: Failed to load config: %v\n", time.Now().Format("15:04:05"), err)
		os.Exit(1)
	}

	//*************** 3. Initialize Core Application Components ************************
	// These are central services like logging, event bus, and GUI.
	mainEventBus := eventbus.New()
	defer mainEventBus.Close()
	mainGui := gui.New(mainCtx, mainEventBus)
	mainLogger, err := logger.NewLogger(mainGui.AddLogMessage, slog.LevelDebug, slog.LevelDebug, slog.LevelDebug, cfg.Paths.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] CRITICAL [main]: Could not create logger: %v\n", time.Now().Format("15:04:05"), err)
		os.Exit(1)
	}
	mainModuleLogger := mainLogger.WithModule("main")

	//*************** 4. Initialize Media Sources (mediadevices) ************************
	mediaSourceManager := mediasource.New(mainCtx, mainLogger, &cfg.MediaSource, mainEventBus)

	//*************** 5. Web Server Component *******************************************
	frontendFS, err := fs.Sub(embeddedFS, "frontend/public")
	if err != nil {
		mainModuleLogger.Error("Failed to create sub-filesystem for frontend", "error", err)
		os.Exit(1)
	}
	mainWebServer := webserver.New(mainCtx, mainLogger, &cfg.WebServer, mainEventBus, frontendFS)

	//*************** 6. OBS Client *****************************************************
	mainObsClient := obsclient.New(mainCtx, mainLogger, &cfg.OBS, mainEventBus)

	//*************** 7. Scheduler (includes internal FileWatcher) **********************
	mainScheduler := scheduler.New(mainCtx, mainLogger, &cfg.Paths, &cfg.Scheduler, mainEventBus)

	//*************** 8. Start Background Services *************************************
	// Launch each long-running service in its own goroutine.
	mainModuleLogger.Info("Starting background runner services...")

	go mediaSourceManager.Run()
	go mainWebServer.Run()
	go mainObsClient.Run()
	go mainScheduler.Run()

	mainModuleLogger.Info("All background services are running.")

	//*************** 9. Start the GUI (Blocking Call) *********************************
	// The GUI runs on the main goroutine and blocks until the user closes it.
	// This is the primary lifetime of the application.
	mainModuleLogger.Info("Starting GUI. This will block until the application exits.")
	mainGui.Run()

	//*************** 10. Shutdown Sequence *********************************************
	// CRITICAL: Stop services in reverse order to prevent race conditions.
	// Services must stop BEFORE the logger/GUI they write to.
	mainModuleLogger.Info("GUI closed. Beginning graceful shutdown sequence.")

	// Step 1: Stop the scheduler (stops generating log messages)
	mainModuleLogger.Info("Stopping scheduler...")
	mainScheduler.Stop()
	time.Sleep(100 * time.Millisecond) // Brief pause to allow final messages to flush

	// Step 2: Stop OBS client
	mainModuleLogger.Info("Stopping OBS client...")
	mainObsClient.Stop()
	time.Sleep(100 * time.Millisecond)

	// Step 3: Stop web server
	mainModuleLogger.Info("Stopping web server...")
	mainWebServer.Stop()
	time.Sleep(100 * time.Millisecond)

	// Step 4: Stop media source manager
	mainModuleLogger.Info("Stopping media source manager...")
	mediaSourceManager.Stop()
	time.Sleep(100 * time.Millisecond)

	// Step 5: Close logger (no more log messages will be accepted)
	mainModuleLogger.Info("Closing logger...")
	mainLogger.Close()

	// Step 6: Event bus closes via defer
	mainModuleLogger.Debug("Application shutdown complete.")
}
