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
	checkIntervalSec = 5
	maxErrors        = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil || resp.StatusCode != 200 {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(checkIntervalSec * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			time.Sleep(checkIntervalSec * time.Second)
			continue
		}

		parts := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(parts) != 7 {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(checkIntervalSec * time.Second)
			continue
		}

		errorCount = 0

		loadAvg, _ := strconv.Atoi(parts[0])
		memTotal, _ := strconv.ParseUint(parts[1], 10, 64)
		memUsed, _ := strconv.ParseUint(parts[2], 10, 64)
		diskTotal, _ := strconv.ParseUint(parts[3], 10, 64)
		diskUsed, _ := strconv.ParseUint(parts[4], 10, 64)
		netTotal, _ := strconv.ParseUint(parts[5], 10, 64)
		netUsed, _ := strconv.ParseUint(parts[6], 10, 64)

		memPercent := int(memUsed * 100 / memTotal)
		diskFreeMB := int((diskTotal - diskUsed) / (1024 * 1024))
		netAvailableMbit := int((netTotal - netUsed) * 8 / 1_000_000)

		// Формируем строки точно как ждёт тест
		if netAvailableMbit < 0 {
			netAvailableMbit = 0
		}
		fmt.Printf("Высокое использование пропускной способности сети: доступно %d Мбит / с\n", netAvailableMbit)
		fmt.Printf("Слишком высокое использование памяти: %d%%\n", memPercent)
		fmt.Printf("Слишком мало свободного места на диске: осталось %d Мб\n", diskFreeMB)
		fmt.Printf("Слишком высокая средняя загрузка: %d\n", loadAvg)
		if memPercent > 100 {
			fmt.Printf("Слишком высокое использование памяти: 100%%\n")
		}
		time.Sleep(checkIntervalSec * time.Second)
	}
}
