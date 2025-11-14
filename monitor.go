package main

import (
	"fmt"
	"io/ioutil"
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
	pollInterval     = 5 * time.Second
	maxFetchErrors   = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= maxFetchErrors {
				fmt.Println("Unable to fetch server statistic.")
				errorCount = 0
			}
			time.Sleep(pollInterval)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= maxFetchErrors {
				fmt.Println("Unable to fetch server statistic.")
				errorCount = 0
			}
			time.Sleep(pollInterval)
			continue
		}

		fields := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= maxFetchErrors {
				fmt.Println("Unable to fetch server statistic.")
				errorCount = 0
			}
			time.Sleep(pollInterval)
			continue
		}

		errorCount = 0 // reset on success

		// Парсим все значения
		loadAvg, _ := strconv.ParseFloat(fields[0], 64)
		totalMem, _ := strconv.ParseFloat(fields[1], 64)
		usedMem, _ := strconv.ParseFloat(fields[2], 64)
		totalDisk, _ := strconv.ParseFloat(fields[3], 64)
		usedDisk, _ := strconv.ParseFloat(fields[4], 64)
		netTotal, _ := strconv.ParseFloat(fields[5], 64)
		netUsed, _ := strconv.ParseFloat(fields[6], 64)

		// Load Average
		if loadAvg > loadThreshold {
			fmt.Printf("Слишком высокая средняя загрузка: %.0f\n", loadAvg)
		}

		// Memory usage
		memPercent := int(usedMem / totalMem * 100)
		if memPercent > int(memoryThreshold*100) {
			fmt.Printf("Слишком высокое использование памяти: %d%%\n", memPercent)
		}

		// Disk free
		diskFree := int((totalDisk - usedDisk) / (1024 * 1024)) // Мб
		if usedDisk/totalDisk > diskThreshold {
			fmt.Printf("Слишком мало свободного места на диске: осталось %d Мб\n", diskFree)
		}

		// Network free in Mbit/s
		netFree := int((netTotal - netUsed) * 8 / 1024 / 1024) // Мбит/с
		if netUsed/netTotal > networkThreshold {
			fmt.Printf("Высокое использование пропускной способности сети: доступно %d Мбит/с\n", netFree)
		}

		time.Sleep(pollInterval)
	}
}
