package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func usage() {
	fmt.Printf("Usage: %s <download-folder> <input-file>\n", os.Args[0])
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func search_and_download(artist string, title string, folder string) {
	repeat_counter := 0
	for repeat_counter <= 1 {
		fmt.Printf("[INFO] Search for title '%s' by artist '%s'\n", title, artist)
		search_query := artist + "+" + title
		search_query = strings.Replace(search_query, " ", "+", -1)
		data := strings.NewReader(fmt.Sprintf("q=%s&page=0", search_query))
		url := "https://myfreemp3juices.cc/api/search.php"
		headers := make(map[string]string)
		headers["accept"] = "text/javascript, application/javascript, application/ecmascript, application/x-ecmascript, */*; q=0.01"
		headers["accept-language"] = "en-US,en;q=0.9"
		headers["content-type"] = "application/x-www-form-urlencoded; charset=UTF-8"
		headers["sec-fetch-dest"] = "empty"
		headers["sec-fetch-mode"] = "cors"
		headers["sec-fetch-site"] = "same-origin"
		headers["sec-gpc"] = "1"
		headers["x-requested-with"] = "XMLHttpRequest"
		headers["cookie"] = "musicLang=en"
		headers["Referer"] = "https://myfreemp3juices.cc/"
		headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
		headers["User-Agent"] = "curl/7.74.0"

		req, err := http.NewRequest("POST", url, data)
		for header, value := range headers {
			req.Header.Set(header, value)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		check(err)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		bodyString := strings.TrimLeft(string(body), "(")
		bodyString = strings.TrimRight(bodyString, ");\n")

		var result map[string]interface{}
		err = json.Unmarshal([]byte(bodyString), &result)
		check(err)

		tracklist, conversion_ok := result["response"].([]interface{})
		if !conversion_ok {
			if repeat_counter <= 2 {
				fmt.Println("   [ERROR] Download failed, trying again in 5 seconds")
				time.Sleep(5 * time.Second)
				repeat_counter += 1
				continue
			} else {
				fmt.Println("   [ERROR] Download failed - try manually")
				return
			}
		}

		mp3url := tracklist[1].(map[string]interface{})["url"]
		out, err := os.Create(filepath.Join(folder, fmt.Sprintf("%s - %s.mp3", artist, title)))
		check(err)
		defer out.Close()
		down_resp, err := http.Get(mp3url.(string))
		check(err)
		defer down_resp.Body.Close()
		fmt.Printf("[INFO] Downloading title '%s' by artist '%s'...\n", title, artist)
		io.Copy(out, down_resp.Body)
		fmt.Println("---")
		repeat_counter = 3
	}
}

func main() {
	if len(os.Args) != 3 {
		usage()
		os.Exit(2)
	}

	folder := os.Args[1]
	input_file := os.Args[2]

	err := os.MkdirAll(folder, 0755)
	check(err)

	dat, err := os.ReadFile(input_file)
	check(err)
	line_by_line := strings.Split(string(dat), "\n")

	for _, line := range line_by_line {
		artist := strings.TrimSpace(strings.Split(line, "|")[0])
		title := strings.TrimSpace(strings.Split(line, "|")[1])
		search_and_download(artist, title, folder)
	}

}
