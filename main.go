package main

import (
	"bufio"
	"github.com/asticode/go-astisub"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var timeLayout = "15:04:05"

type message struct {
	time time.Duration
	text string
}

func readMessagesFromReader(reader io.Reader) []message {
	scanner := bufio.NewScanner(reader)
	var startTime time.Time

	var messages []message

	first := true
	for scanner.Scan() {
		line := scanner.Text()
		absoluteTime, err := time.Parse(timeLayout, line[:len(timeLayout)])

		if err != nil {
			log.Fatal(err)
		}

		if first {
			first = false
			startTime = absoluteTime
		}

		messages = append(messages, message{
			time: absoluteTime.Sub(startTime),
			text: line[len(timeLayout)+2:],
		})
	}

	return messages
}

func readMessages(chatFilePath string) ([]message, error) {
	chatFile, err := os.Open(chatFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = chatFile.Close()
	}()

	return readMessagesFromReader(chatFile), nil
}

func getSubs(messages []message) *astisub.Subtitles {
	subs := astisub.NewSubtitles()

	for i, _ := range messages[:len(messages)-1] {
		subs.Items = append(subs.Items, &astisub.Item{
			Comments:    nil,
			EndAt:       messages[i+1].time,
			InlineStyle: nil,
			Lines: []astisub.Line{
				{
					Items: []astisub.LineItem{
						{InlineStyle: nil, Style: nil, Text: messages[i].text},
					},
					VoiceName: "",
				},
			},
			Region:  nil,
			StartAt: messages[i].time,
			Style:   nil,
		})
	}

	return subs
}

func writeSubs(subs *astisub.Subtitles, srtFileName string) error {
	srtFile, err := os.OpenFile(srtFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer srtFile.Close()

	err = subs.WriteToSRT(srtFile)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("No chat file provided")
	}

	chatFilePath := os.Args[1]
	messages, err := readMessages(chatFilePath)
	if err != nil {
		log.Fatal(err)
	}

	subs := getSubs(messages)

	chatFileName := filepath.Base(chatFilePath)
	srtFileName := strings.TrimSuffix(chatFileName, filepath.Ext(chatFileName)) + ".srt"

	err = writeSubs(subs, srtFileName)
	if err != nil {
		log.Fatal(err)
	}
}
