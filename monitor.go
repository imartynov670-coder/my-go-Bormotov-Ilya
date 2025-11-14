package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Симуляция ответов сервера (как если бы мы опрашивали http://srv.msk01.gigacorp.local/_stats)
var mockResponses = []string{
	"25,2147483648,1800000000,5497558138880,4940000000000,104857600,95000000",
	"35,2147483648,1900000000,5497558138880,5000000000000,104857600,100000000",
	"28,2147483648,1600000000,5497558138880,4500000000000,104857600,90000000",
	"40,2147483648,1800000000,5497558138880,5200000000000,104857600,95000000",
	"22,2147483648,1700000000,5497558138880,4000000000000,104857600,85000000",
	"31,2147483648,2000000000,5497558138880,5300000000000,104857600,102000000",
	"29,2147483648,1800000000,5497558138880,4800000000000,104857600,91000000",
}

func main() {
	errorCount := 0

	for i := 0; i < len(mockResponses); i++ {
		line := mockResponses[i]
		fields := strings.Split(line, ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			continue
		}

		loadAvg, _ := strconv.Atoi(fields[0])
		memTotal, _ := strconv.ParseInt(fields[1], 10, 64)
		memUsed, _ := strconv.ParseInt(fields[2], 10, 64)
		diskTotal, _ := strconv.ParseInt(fields[3], 10, 64)
		diskUsed, _ := strconv.ParseInt(fields[4], 10, 64)
		netTotal, _ := strconv.ParseInt(fields[5], 10, 64)
		netUsed, _ := strconv.ParseInt(fields[6], 10, 64)

		// Проверяем пороги
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %d\n", loadAvg)
		}

		memPerc := int(float64(memUsed) / float64(memTotal) * 100)
		if memPerc > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memPerc)
		}

		diskFreeMb := int((diskTotal - diskUsed) / 1024 / 1024)
		if float64(diskUsed)/float64(diskTotal) > 0.9 {
			fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMb)
		}

		netFreeMbit := int(float64(netTotal-netUsed) * 8 / 1000000)
		if float64(netUsed)/float64(netTotal) > 0.9 {
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netFreeMbit)
		}

		time.Sleep(1 * time.Second) // имитация задержки между запросами
	}
}

