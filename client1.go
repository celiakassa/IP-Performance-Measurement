package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var url = flag.String("url", "https://monitor.uac.bj:4449/config", "Give a good url")

const (
	smallContentLength = 1
	largeContentLength = 4 * 1024 * 1024 * 1024
	chunkSize          = 64 * 1024
)

var (
	buffed []byte
)

func init() {
	buffed = make([]byte, chunkSize)
	for i := range buffed {
		buffed[i] = 'x'
	}
}

func cli(i int) {
	client1 := http.Client{
		Timeout: time.Duration(30) * time.Second,
	}
	resp1, err1 := client1.Get(*url)
	fmt.Printf("connexion numero %T  reussie \n", i)
	if err1 != nil {
		fmt.Printf("Error:  %s", err1)
		return
	}
	defer resp1.Body.Close()
	//config1, err1 := ioutil.ReadAll(resp1.Body)

}

type Url struct {
	Small_https_download_url string `json:"small_https_download_url"`
	Large_https_download_url string `json:"large_https_download_url"`
	Https_upload_url         string `json:"https_upload_url"`
}
type Config struct {
	Version int `json:"version"`
	Urls    Url `json:"urls"`
}


func main() {
	flag.Parse()
	client := http.Client{
		Timeout: time.Duration(30) * time.Second,
	}
	resp, err := client.Get(*url)
	if err != nil {
		fmt.Printf("Error:  %s", err)
		return
	}
	defer resp.Body.Close() 
	config, err := ioutil.ReadAll(resp.Body)

	for i := 0; i < 4; i++ {
		go cli(i)
	}
	n :=0
	ticker := time.NewTicker(1 *time.Second)
	for t := range ticker.C {
	 fmt.Println("Invoked at",t)
	 n+=1
	 if n==15 {
	 	break
	 	}
	}
	var conf Config
	json.Unmarshal(config, &conf)

	fmt.Println("Config:", conf.Urls)

//download
	links := conf.Urls
	t1 := time.Now()
	resp, errD := client.Get(links.Large_https_download_url)
	if errD != nil {
		fmt.Printf("Error:  %s", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	aff := struct {
		Duration time.Duration `json:"DurationMS"`
		Bytes    int64
		BPS      string `json:"Download Speed"`
	}{
		Duration: time.Since(t1) / time.Millisecond,
		Bytes:    int64(len(body)),
	}

	if aff.Duration > 0 && aff.Bytes > 0 {
		sie := ((float64(aff.Bytes) * 8) / (float64(aff.Duration) / 1000)) / 1000000
		aff.BPS = strconv.FormatInt(int64(sie), 10) + " Mbps"
	}

	js, err := json.Marshal(aff)
	if err != nil {
		fmt.Printf("Error:  %s", err)
		return
	}
	fmt.Println("Stats:")
	fmt.Println(string(js))

	//  Upload
	responseBody := bytes.NewBuffer(buffed)
	//Leverage Go's HTTP Post function to make request
	respU, errU := http.Post(links.Https_upload_url, "application/json", responseBody)
	if errU != nil {
		fmt.Printf("Error:  %s", errU)
		return
	}
	defer respU.Body.Close()
	bodyU, _ := ioutil.ReadAll(respU.Body)

	fmt.Println("Stats:")
	fmt.Println(string(bodyU))

}
