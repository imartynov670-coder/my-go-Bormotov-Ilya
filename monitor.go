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
	url           = "http://srv.msk01.gigacorp.local/_stats"
	loadLimit     = 30
	memUsageLimit = 0.8
	diskUsageLimit = 0.9
	netUsageLimit  = 0.9
	pollInterval  = 2 * time.Second
	maxErrors     = 3
)

func fetchStats() ([]int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data length")
	}
	nums := make([]int64, 7)
	for i, p := range parts {
		nums[i], err = strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %v", err)
		}
	}
	return nums, nil
}

func checkAndPrint(stats []int64) {
	load := stats[0]
	totalMem := stats[1]
	usedMem := stats[2]
	totalDisk := stats[3]
	usedDisk := stats[4]
	netTotal := stats[5]
	netUsed := stats[6]

	// Load
	if load > loadLimit {
		fmt.Printf("Слишком высокая средняя загрузка: %d\n", load)
	}

	// Memory
	memPercent := int(float64(usedMem) / float64(totalMem) * 100)
	if float64(usedMem)/float64(totalMem) > memUsageLimit {
		fmt.Printf("Слишком высокое использование памяти: %d%%\n", memPercent)
	}

	// Disk
	freeDisk := totalDisk - usedDisk
	if float64(usedDisk)/float64(totalDisk) > diskUsageLimit {
		fmt.Printf("Слишком мало свободного места на диске: осталось %d Мб\n", freeDisk/1024/1024)
	}

	// Network
	freeNet := netTotal - netUsed
	if float64(netUsed)/float64(netTotal) > netUsageLimit {
		// Преобразуем в Мбит/с
		mbit := freeNet * 8 / 1000 / 1000
		fmt.Printf("Высокая загрузка сети: доступно %d Мбит/с\n", mbit)
	}
}

func main() {
	errorCount := 0
	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
			}
		} else {
			errorCount = 0
			checkAndPrint(stats)
		}
		time.Sleep(pollInterval)
	}
}
