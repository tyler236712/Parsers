package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Subtitle struct {
	Index     int
	StartTime time.Duration
	EndTime   time.Duration
	Text      []string
}

func main() {
	filePath := "./srt/test-JP.srt"

	subtitiles, err := parseSRTFile(filePath)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	for _, subtitle := range subtitiles {
		fmt.Printf("%+v\n", subtitle)
	}
}

func parseTimestamp(timestamp string) (time.Duration, error) {
	var hr, min, sec, ms int
	_, err := fmt.Fscanf(strings.NewReader(timestamp), "%02d:%02d:%02d,%03d", &hr, &min, &sec, &ms)
	if err != nil {
		return 0, err
	}

	duration := time.Duration(hr)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second + time.Duration(ms)*time.Millisecond

	return duration, nil
}

func parseSRTFile(filePath string) ([]Subtitle, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var subtitles []Subtitle
	var curr *Subtitle
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// blank line signifies start of new subtitle
		if line == "" && curr != nil {
			subtitles = append(subtitles, *curr)
			curr = nil
			continue
		}

		// first comes the index
		if curr == nil {
			curr = new(Subtitle)
			_, err = fmt.Fscanf(strings.NewReader(line), "%d", &curr.Index)
			if err != nil {
				return nil, err
			}
			continue
		}

		// next comes the timestamps
		if curr.StartTime == 0 || curr.EndTime == 0 {
			timeRange := strings.Split(line, " --> ")
			if len(timeRange) != 2 {
				return nil, errors.New("Invalid time format found while parsing.")
			}

			start, err := parseTimestamp(timeRange[0])
			if err != nil {
				return nil, err
			}

			end, err := parseTimestamp(timeRange[1])
			if err != nil {
				return nil, err
			}

			curr.StartTime = start
			curr.EndTime = end

			continue
		}

		curr.Text = append(curr.Text, line)

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subtitles, err
}
