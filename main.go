package main

import (
	"fmt"
	"io"
	"io/ioutil"
	// "math"
	"net/http"
	"os"
	// "regexp"
	"strconv"
	// "time"
)

var BaseUrl = "http://downloads.openwrt.org.cn/PandoraBox/ralink/packages/"
var Package = "telephony"

var mq = make(chan Ipk)
var exitChan = make(chan int)

type Ipk struct {
	data     io.ReadCloser
	filename string
	url      string
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func Mkdir(filename string) {
	flag := Exist(filename)
	if !flag {
		os.Mkdir(filename, 0755)
	}
}

func Get(url string) (content string, statusCode int) {
	resp, err := http.Get(url)
	if err != nil {
		statusCode = -100
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		statusCode = -200
		return
	}
	statusCode = resp.StatusCode
	content = string(data)
	return
}

func download(ipks []Ipk, index int) {
	for k, ipk := range ipks {
		if ipk.url != "" {
			if Exist("./Download/" + Package + "/" + ipk.filename) {
				fmt.Println(ipk.filename + " has been downloaded, passed")
				continue
			}
			res, _ := http.Get(ipk.url)
			ipk.data = res.Body
			fmt.Println(res.Status)
			mq <- ipk
			fmt.Println("download success No." + strconv.Itoa(index+k) + ":" + ipk.filename)
			// time.Sleep(0.5)
		}
	}
	exitChan <- 1
}

func receiver() {
	for {
		ipk := <-mq
		file, err := os.Create("./Download/" + Package + "/" + ipk.filename)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(file, ipk.data)
		if err != nil {
			panic(err)
		}
		println("write success: " + ipk.filename)
		file.Close()
		ipk.data.Close()
	}
}

func main() {
	// http.Handle("/pandorabox/", http.StripPrefix("/pandorabox/", http.FileServer(http.Dir("./Download"))))
	// http.ListenAndServe(":9090", nil)

	Mkdir("Download")
	Mkdir("Download/" + Package)
	html, _ := Get(BaseUrl + Package)
	re := regexp.MustCompile(`<a href="(.*?)">`)
	list := re.FindAllStringSubmatch(html, -1)
	slice := make([]Ipk, 10)
	for k, v := range list {
		if v[1] != "../" {
			fmt.Println(strconv.Itoa(k) + ":" + v[1])
			ipk := Ipk{data: nil, filename: v[1], url: BaseUrl + Package + "/" + v[1]}
			slice = append(slice, ipk)
		}
	}
	var mid = int(math.Floor(float64(len(slice)) / float64(2)))
	go download(slice[:mid], 0)
	go download(slice[mid+1:], mid+1)
	go receiver()
	go receiver()

	var a = 0
	for {
		select {
		case b := <-exitChan:
			a += b
		}
		if a == 2 {
			break
		}
	}
	fmt.Println("exit")
}
