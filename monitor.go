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
	pollInterval     = 5 * time.Second
	maxErrorCount    = 3
	loadThreshold    = 30
	memUsagePercent  = 80
	diskUsagePercent = 90
	netUsagePercent  = 90
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil || resp.StatusCode != 200 {
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(pollInterval)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			time.Sleep(pollInterval)
			continue
		}

		values := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(values) != 7 {
			errorCount++
			time.Sleep(pollInterval)
			continue
		}

		errorCount = 0 // сброс ошибок при успешном получении

		// парсим целые значения
		load, _ := strconv.ParseInt(values[0], 10, 64)
		totalMem, _ := strconv.ParseInt(values[1], 10, 64)
		usedMem, _ := strconv.ParseInt(values[2], 10, 64)
		totalDisk, _ := strconv.ParseInt(values[3], 10, 64)
		usedDisk, _ := strconv.ParseInt(values[4], 10, 64)
		totalNet, _ := strconv.ParseInt(values[5], 10, 64)
		usedNet, _ := strconv.ParseInt(values[6], 10, 64)

		// собираем все сообщения в слайс
		var msgs []string

		if load > loadThreshold {
			msgs = append(msgs, fmt.Sprintf("Слишком высокая средняя загрузка: %d", load))
		}

		memPercent := int(usedMem * 100 / totalMem)
		if memPercent > memUsagePercent {
			msgs = append(msgs, fmt.Sprintf("Слишком высокое использование памяти: %d%%", memPercent))
		}

		freeDisk := totalDisk - usedDisk
		if usedDisk*100/totalDisk > diskUsagePercent {
			msgs = append(msgs, fmt.Sprintf("Слишком мало свободного места на диске: осталось %d Мб", freeDisk/1024/1024))
		}

		freeNet := totalNet - usedNet
		if usedNet*100/totalNet > netUsagePercent {
			msgs = append(msgs, fmt.Sprintf("Высокая загрузка сети: доступно %d Мбит/с", freeNet*8/1024/1024))
		}

		// выводим ровно 7 сообщений (или меньше, если их меньше)
		for i, msg := range msgs {
			if i >= 7 {
				break
			}
			fmt.Println(msg)
		}

		time.Sleep(pollInterval)
	}
}
