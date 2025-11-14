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
	memoryThreshold  = 80.0
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

func fetchStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}

	line := strings.TrimSpace(scanner.Text())
	parts := strings.Split(line, ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format")
	}

	stats := make([]float64, 7)
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %v", err)
		}
		stats[i] = val
	}

	return stats, nil
}

func checkMetrics(s []float64) {
	load := s[0]
	totalMem, usedMem := s[1], s[2]
	totalDisk, usedDisk := s[3], s[4]
	totalNet, usedNet := s[5], s[6]

	// 1. Load Average
	if load > loadThreshold {
		fmt.Printf("Слишком высокая средняя загрузка: %.0f\n", load)
	}

	// 2. Memory usage
	if totalMem > 0 {
		memPercent := math.Floor((usedMem / totalMem) * 100)
		if memPercent > memoryThreshold {
			fmt.Printf("Слишком высокое использование памяти: %.0f%%\n", memPercent)
		}
	}

	// 3. Disk usage
	if totalDisk > 0 && usedDisk/totalDisk > diskThreshold {
		freeMb := math.Floor((totalDisk - usedDisk) / (1024 * 1024))
		fmt.Printf("Слишком мало свободного места на диске: осталось %.0f Мб\n", freeMb)
	}

	// 4. Network usage
	if totalNet > 0 && usedNet/totalNet > networkThreshold {
		freeMbit := math.Floor((totalNet - usedNet) / 1_000_000)
		fmt.Printf("Высокая загрузка сети: доступно %.0f Мбит/с\n", freeMbit)
	}
}
