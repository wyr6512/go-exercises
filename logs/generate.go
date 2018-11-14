package main

import (
	"flag"
	"fmt"
	// "io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type resource struct {
	url    string
	target string
	start  int
	end    int
}

func ruleResource() []resource {
	var res []resource
	r1 := resource{
		url:    "http://localhost:8088/",
		target: "",
		start:  0,
		end:    0,
	}
	r2 := resource{
		url:    "http://localhost:8088/list/{$id}.html",
		target: "{$id}",
		start:  1,
		end:    21,
	}
	r3 := resource{
		url:    "http://localhost:8088/detail/{$id}.html",
		target: "{$id}",
		start:  1,
		end:    12924,
	}
	res = append(append(append(res, r1), r2), r3)
	return res
}

func buildUrl(res []resource) []string {
	var list []string
	for _, r := range res {
		if len(r.target) == 0 {
			list = append(list, r.url)
		} else {
			for i := r.start; i <= r.end; i++ {
				url := strings.Replace(r.url, r.target, strconv.Itoa(i), -1)
				list = append(list, url)
			}
		}
	}
	return list
}

//模拟日志
func makeLog(current, refer, ua string) string {
	u := url.Values{}
	u.Set("time", strconv.FormatInt(time.Now().Unix(), 10))
	u.Set("url", current)
	u.Set("refer", refer)
	u.Set("ua", ua)
	paramsStr := u.Encode()

	logTemplate := "127.0.0.1 - - [13/Nov/2018:06:32:00 +0000] \"GET /dig?{$paramsStr} HTTP/2.0\" 200 127 \"-\" \"{$ua}}\""
	log := strings.Replace(logTemplate, "{$paramsStr}", paramsStr, -1)
	log = strings.Replace(log, "{$ua}", ua, -1)
	return log
}

//useragent
var uaList = []string{
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.1 (KHTML, like Gecko) Chrome/14.0.835.163 Safari/535.1",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Opera/9.80 (Windows NT 6.1; U; zh-cn) Presto/2.9.168 Version/11.50",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 2.0.50727; SLCC2; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; InfoPath.3; .NET4.0C; Tablet PC 2.0; .NET4.0E)",
}

//随机
func randInt(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if min > max {
		return max
	}
	return r.Intn(max-min) + min
}
func main() {
	total := flag.Int("total", 100, "how many rows by created")
	filePath := flag.String("filePath", "/e/goproj/src/go-exercises/logs/access.log", "usage")
	flag.Parse()

	res := ruleResource()
	list := buildUrl(res)
	var logStr string
	for i := 0; i < *total; i++ {
		currentUrl := list[randInt(0, len(list)-1)]
		referUrl := list[randInt(0, len(list)-1)]
		ua := uaList[randInt(0, len(uaList)-1)]
		logStr += makeLog(currentUrl, referUrl, ua) + "\n"
		// ioutil.WriteFile(*filePath, []byte(logStr), 0644)
	}
	fd, err := os.OpenFile(*filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	fd.Write([]byte(logStr))
	defer fd.Close()
	fmt.Println(*total, *filePath)
}
