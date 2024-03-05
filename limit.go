package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	filename = "openai-text-service-request-timestamps"
)

var timestamps []int64
var limitPerDayLock sync.Mutex

func probeLimitPerDay() bool {
	limitPerDayLock.Lock()
	defer limitPerDayLock.Unlock()
	now := time.Now().Unix()
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	sliceIndex := 0

	for k, timestamp := range timestamps {
		if timestamp < yesterday {
			sliceIndex = k + 1
		} else {
			break
		}
	}
	timestamps = timestamps[sliceIndex:]

	if len(timestamps) < limit {
		timestamps = append(timestamps, now)
		return true
	}
	return false
}

func readTimestampsFromFile() {
	file, err := os.Open(os.TempDir() + filename)
	if err != nil {
		return
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		timestamp, err := strconv.Atoi(fileScanner.Text())
		if err == nil {
			timestamps = append(timestamps, int64(timestamp))
		}
	}
}

func timestampsWriter(writeInterval time.Duration) {
	ticker := time.NewTicker(writeInterval)

	go func() {
		for range ticker.C {
			fmt.Println("writing to ", os.TempDir()+filename, timestamps)
			file, err := os.OpenFile(os.TempDir()+filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				panic(err)
			}
			writer := bufio.NewWriter(file)
			for _, timestamp := range timestamps {
				writer.WriteString(strconv.Itoa(int(timestamp)))
				writer.WriteString("\n")
			}
			writer.Flush()
			file.Close()
		}
	}()
}

func initDailyLimit() {
	readTimestampsFromFile()
	timestampsWriter(time.Minute)
}
