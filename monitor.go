package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL        = "http://srv.msk01.gigacorp.local/_stats"
	loadThreshold    = 30.0
	memoryThreshold  = 0.80
	diskThreshold    = 0.90
	networkThreshold = 0.90
	maxErrors        = 3
	checkInterval    = 30 * time.Second
)

func main() {
	errorCount := 0

	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(checkInterval)
			continue
		}

		errorCount = 0
		checkMetrics(stats)
		time.Sleep(checkInterval)
	}
}

// Запрос данных у сервера
func fetchStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status")
	}

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}

	line := strings.TrimSpace(scanner.Text())
	parts := strings.Split(line, ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid fields count")
	}

	stats := make([]float64, 7)
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number")
		}
		stats[i] = val
	}

	return stats, nil
}

// Проверка метрик и вывод сообщений
func checkMetrics(stats []float64) {
	// 1. Load Average
	loadAvg := stats[0]
	if loadAvg > loadThreshold {
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

	// 2. Memory usage
	totalMem := stats[1]
	usedMem := stats[2]
	if totalMem > 0 {
		usage := (usedMem / totalMem) * 100
		if usage > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", usage)
		}
	}

	// 3. Disk usage
	totalDisk := stats[3]
	usedDisk := stats[4]
	if totalDisk > 0 {
		usage := usedDisk / totalDisk
		if usage > diskThreshold {
			freeMB := math.Floor((totalDisk - usedDisk) / (1024 * 1024))
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeMB)
		}
	}

	// 4. Network bandwidth
	totalNet := stats[5]
	usedNet := stats[6]
	if totalNet > 0 {
		usage := usedNet / totalNet
		if usage > networkThreshold {
			// свободная пропускная способность в мегабитах
			freeMbit := (totalNet - usedNet) * 8 / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}
	}
}
