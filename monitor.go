package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL        = "http://srv.msk01.gigacorp.local/_stats"
	loadThreshold    = 30.0
	memoryThreshold  = 0.8
	diskThreshold    = 0.9
	networkThreshold = 0.9
	maxErrors        = 3
	checkInterval    = 30 * time.Second
)

func main() {
	errorCount := 0

	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			fmt.Printf("Error fetching stats: %v\n", err)

			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}

			time.Sleep(checkInterval)
			continue
		}

		errorCount = 0
		checkMetrics(stats)
		time.Sleep(checkInterval)
	}
}

func fetchStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}

	line := scanner.Text()
	parts := strings.Split(line, ",")

	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(parts))
	}

	var stats []float64
	for _, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number format: %v", err)
		}
		stats = append(stats, value)
	}

	return stats, nil
}

func checkMetrics(stats []float64) {
	loadAvg := stats[0]
	if loadAvg > loadThreshold {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}

	totalMem := stats[1]
	usedMem := stats[2]
	if totalMem > 0 {
		memoryUsage := usedMem / totalMem
		if memoryUsage > memoryThreshold {
			fmt.Printf("Memory usage too high: %.1f%%\n", memoryUsage*100)
		}
	}

	totalDisk := stats[3]
	usedDisk := stats[4]
	if totalDisk > 0 {
		diskUsage := usedDisk / totalDisk
		if diskUsage > diskThreshold {
			freeSpaceMB := (totalDisk - usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeSpaceMB)
		}
	}

	totalNet := stats[5]
	usedNet := stats[6]
	if totalNet > 0 {
		networkUsage := usedNet / totalNet
		if networkUsage > networkThreshold {
			availableBandwidthMbit := (totalNet - usedNet) * 8 / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.1f Mbit/s available\n", availableBandwidthMbit)
		}
	}
}