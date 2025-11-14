func checkMetrics(stats []float64) {
    loadAvg := stats[0]
    if loadAvg > 30 {
        fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
    }

    totalMem := stats[1]
    usedMem := stats[2]
    if totalMem > 0 {
        memoryUsage := usedMem / totalMem * 100
        if memoryUsage > 80 {
            fmt.Printf("Memory usage too high: %.0f%%\n", memoryUsage)
        }
    }

    totalDisk := stats[3]
    usedDisk := stats[4]
    if totalDisk > 0 {
        diskUsage := usedDisk / totalDisk
        if diskUsage > 0.9 {
            freeMB := (totalDisk - usedDisk) / (1024 * 1024)
            fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeMB)
        }
    }

    totalNet := stats[5]
    usedNet := stats[6]
    if totalNet > 0 {
        netUsage := usedNet / totalNet
        if netUsage > 0.9 {
            // По ТЗ нужно считать в мегабитах:
            // (байты/сек) * 8 / 1_000_000
            freeMbit := (totalNet - usedNet) * 8 / 1_000_000
            fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
        }
    }
}
