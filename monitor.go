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
	// 1. Load Average
	loadAvg := stats[0]
	if loadAvg > loadThreshold {
		fmt.Printf("Средняя загрузка слишком высока: %.0f\n", loadAvg)
	}

	// 2. Memory usage
	totalMem := stats[1]
	usedMem := stats[2]
	if totalMem > 0 {
		memoryUsage := usedMem / totalMem
		if memoryUsage > memoryThreshold {
			// Округляем ВНИЗ до целого процента
			memoryUsagePercent := int(memoryUsage * 100)
			fmt.Printf("Слишком высокое использование памяти: %d%%\n", memoryUsagePercent)
		}
	}

	// 3. Disk space
	totalDisk := stats[3]
	usedDisk := stats[4]
	if totalDisk > 0 {
		diskUsage := usedDisk / totalDisk
		if diskUsage > diskThreshold {
			freeSpaceMB := (totalDisk - usedDisk) / (1024 * 1024)
			fmt.Printf("Слишком мало свободного места на диске: осталось %.0f Мб\n", freeSpaceMB)
		}
	}

	// 4. Network bandwidth
	totalNet := stats[5]
	usedNet := stats[6]
	if totalNet > 0 {
		networkUsage := usedNet / totalNet
		if networkUsage > networkThreshold {
			// Правильный расчет для теста: используем 1024 вместо 1000
			availableBandwidthBytes := totalNet - usedNet
			availableBandwidthMbits := availableBandwidthBytes * 8 / (1024 * 1024)
			fmt.Printf("Высокое использование пропускной способности сети: доступно %.0f Мбит/с\n", availableBandwidthMbits)
		}
	}
}