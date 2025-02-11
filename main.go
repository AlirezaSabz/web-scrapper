package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

func main() {
	var wg sync.WaitGroup
	var pageLink string
	fmt.Println(" Insert Website Address   : ")
	fmt.Scanln(&pageLink)
	// defaultPageLink :="https://castbox.fm/episode/The-Bonus-Christmas-Round-with-Alfie-Boe-%26-Michael-Ball-id6285457-id764035284?country=gb"

	// create a collector and request to collect information
	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML("audio source", func(e *colly.HTMLElement) {

		link := e.Attr("src")

		if strings.Contains(link, ".mp3") {
			fileSize, err := getFileSize(link)
			if err != nil {
				fmt.Println("ERROR In Getting File Size : ", err)
				return
			}
			parts := make([][]byte, 5)
			e.Request.Visit(link)
			fmt.Println("File Size : ", fileSize)

			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					err := downloadPart(link, i, int(fileSize), &parts[i], &wg)
					if err != nil {
						fmt.Println("ERROR download Part ", i+1, " :", err)
						return
					}
				}()
			}
			wg.Wait()

			// Saving File
			err = saveToFile(link, parts)
			if err != nil {
				fmt.Println("ERROR Saving File: ", err)
				return
			} else {
				fmt.Println("Successfully Saved!!")
			}

		} else {
			fmt.Println("Oops! We Didnt Found Your Music :( ")
			return
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Request To  :   ", r.URL)
	})

	c.Visit(pageLink)

}

func downloadPart(link string, partIndex int, fileSize int, part *[]byte, wg *sync.WaitGroup) error {
	partSize := fileSize / 5

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return err
	}
	start := partIndex * partSize
	end := (partIndex+1)*partSize - 1
	if partIndex == 4 {
		end = fileSize - 1
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// var buf *bytes.Buffer
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {

		fmt.Println("ERROR In Copy Response ", partIndex+1, err)
		return err
	}

	*part = buf.Bytes()
	fmt.Printf(" Section %d Downloaded !  \n", partIndex+1)
	defer wg.Done()
	return nil

}

func getFileSize(mp3URL string) (int64, error) {

	resp, err := http.Head(mp3URL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// fmt.Println("************************")
	// fmt.Println("response headers : ")
	// for key, value := range resp.Header {
	// 	fmt.Printf("%s: %s\n", key, value)
	// }
	// fmt.Println("************************")

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("Content-Length not found")
	}

	//convert string to int
	var fileSize int64
	fmt.Sscanf(contentLength, "%d", &fileSize)

	return fileSize, nil
}

func saveToFile(link string, parts [][]byte) error {
	// create file
	fileName := path.Base(link)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// save downloaded data in to the file
	for i := 0; i < 5; i++ {
		_, err := file.Write(parts[i])
		if err != nil {
			return err
		}
	}
	return nil
}
