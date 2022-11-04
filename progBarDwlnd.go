package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	download := NewDownload("https://download.mozilla.org/?product=firefox-latest-ssl&os=osx&lang=ru",
		"firefox.dmg")
	progressBar := ProgressBar{download}
	progressBar.Start()
}

type Download struct {
	File          *os.File
	Response      *http.Response
	ContentLength int
	Done          bool
}

func (download *Download) StartDownload() {
	_, err := io.Copy(download.File, download.Response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		download.Done = true
		return
	}
	download.Done = true
}

func NewDownload(url string, fileName string) *Download {
	// выделяем память для нового объекта new(Download)
	download := new(Download)

	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error:", err)
	}
	download.File = out

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
	}

	download.Response = response
	download.ContentLength, _ = strconv.Atoi(response.Header.Get("content-length"))
	download.Done = false

	return download
}

func (download *Download) BytesDownloaded() int {
	info, err := download.File.Stat()
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}
	return int(info.Size())
}

type ProgressBar struct {
	Download *Download
}

func (progressBar *ProgressBar) Start() {
	// запускаем загрузку в горутине
	go progressBar.Download.StartDownload()
	progressBar.Show()

	// закрываем файл и коннект
	progressBar.Download.Response.Body.Close()
	progressBar.Download.File.Close()
}

func (progressBar *ProgressBar) Show() {
	var progress int
	totalBytes := int(progressBar.Download.ContentLength)
	lastTime := false

	// пока идет загрузка файла рисуем прогресс бар
	for !progressBar.Download.Done || lastTime {
		// Start the progress bar - carriage return to overwrite previous iteration
		fmt.Print("\r[")
		bytesDone := progressBar.Download.BytesDownloaded()
		progress = 40 * bytesDone / totalBytes

		// рисуем прогресс бар
		for i := 0; i < 40; i++ {
			if i < progress {
				fmt.Print("=")
			} else if i == progress {
				fmt.Print(">")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("] ")
		fmt.Printf("%d/%dkB", bytesDone/1000, totalBytes/1000)
		time.Sleep(100 * time.Millisecond)

		// после загрузки еще одна пробежка по циклу
		if progressBar.Download.Done && !lastTime {
			lastTime = true
		} else {
			lastTime = false
		}
	}
	fmt.Println()
}
