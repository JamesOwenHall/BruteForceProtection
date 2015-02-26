package main

import (
	"flag"
	"fmt"
	"github.com/JamesOwenHall/BruteForceProtection/core/config"
	"github.com/JamesOwenHall/BruteForceProtection/core/dashboard"
	"github.com/JamesOwenHall/BruteForceProtection/core/hitcounter"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	// Setup multithreading
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse flags
	configFilename := flag.String("c", "config.json", "the name of the configuration file")
	displayVersion := flag.Bool("version", false, "display the version number")
	flag.Parse()

	// Display version number
	if *displayVersion {
		fmt.Println("BFP core version", config.Version)
		fmt.Println("Copyright (C) James Hall 2015.")
		return
	}

	// Read the configuration
	configuration, errs := config.ReadConfig(*configFilename)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, "configuration error:", err)
		}

		return
	}

	// Create server
	counter := hitcounter.NewHitCounter(configuration.Directions)
	defer counter.Close()

	// Start message server
	err := counter.ListenAndServe(configuration.ListenType, configuration.ListenAddress)
	if err == nil {
		fmt.Println("Listening for hits @", configuration.ListenAddress)
	} else {
		fmt.Println("Server error: can't listen @", configuration.ListenAddress)
		return
	}

	// Start the dashboard server
	var dash *dashboard.Server
	if configuration.DashboardAddress != "" {
		dash = dashboard.New(configuration, counter)
		dash.ListenAndServe()
		fmt.Println("Dashboard listening @", configuration.DashboardAddress)
	}

	// Capture interrupt signal so that the server closes properly
	interrupts := make(chan os.Signal)
	signal.Notify(interrupts, os.Interrupt)
	<-interrupts
}
