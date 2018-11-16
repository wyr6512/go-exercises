package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mgutz/str"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	HANDLE_DIG    = " /dig?"
	HANDLE_DETAIL = "/detail/"
	HANDLE_LIST   = "/list/"
	HANDLE_HTML   = ".html"
)

type cmdParams struct {
	logFilePath string
	routineNum  int
}
type digData struct {
	time  string
	url   string
	refer string
	ua    string
}

type urlData struct {
	data  digData
	uid   string
	unode urlNode
}

type urlNode struct {
	unType string //详情页列表页首页
	unRid  int    //resource id
	unUrl  string
	unTime string
}

type storageBlock struct {
	counterType  string
	storageModel string
	unode        urlNode
}

var log = logrus.New()

func init() {
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	logFilePath := flag.String("logFilePath", "/e/goproj/src/go-exercises/logs/access.log", "usage")
	routineNum := flag.Int("routine", 5, "usage")
	l := flag.String("l", "/e/goproj/src/go-exercises/logs/error.log", "usage")
	flag.Parse()
	params := cmdParams{
		*logFilePath,
		*routineNum,
	}

	logFd, err := os.OpenFile(*l, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Out = logFd
		defer logFd.Close()
	}
	log.Infof("exec start")
	log.Infof("params: logFilePath=%s, routineNum=%d, logPath=%s", params.logFilePath, params.routineNum, *l)

	redisPool, err := pool.New("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		log.Fatalln("redis pool created failed")
		panic(err)
	} else {
		go func() {
			for {
				redisPool.Cmd("PING")
				time.Sleep(3 * time.Second)
			}
		}()
	}
	var logChannel = make(chan string, 3*params.routineNum)
	var pvChannel = make(chan urlData, params.routineNum)
	var uvChannel = make(chan urlData, params.routineNum)
	var storageChannel = make(chan storageBlock, params.routineNum)

	//日志消费者，逐行读取日志文件放入logchannel
	go readFileLinebyLine(params, logChannel)

	//创建一组日志处理，处理日志格式化放入pvchannel和uvchannel
	for i := 0; i < params.routineNum; i++ {
		go logConsumer(logChannel, pvChannel, uvChannel)
	}

	//创建pv uv统计放入storagechannel
	go pvCounter(pvChannel, storageChannel)
	go uvCounter(uvChannel, storageChannel, redisPool)

	//结果存储
	go dataStorage(storageChannel, redisPool)
}

//redis存储统计结果，实际场景用HBase
func dataStorage(storageChannel chan storageBlock, redisPool *pool.Pool) {
	for block := range storageChannel {
		prefix := block.counterType + "-"

		setKeys := []string{
			prefix + "day_" + getTime(block.unode.unTime, "day"),
			prefix + "hour_" + getTime(block.unode.unTime, "hour"),
			prefix + "min_" + getTime(block.unode.unTime, "min"),
			prefix + block.unode.unType + "_day_" + getTime(block.unode.unTime, "day"),
			prefix + block.unode.unType + "_hour_" + getTime(block.unode.unTime, "hour"),
			prefix + block.unode.unType + "_min_" + getTime(block.unode.unTime, "min"),
		}

		rowId := block.unode.unRid
		for _, key := range setKeys {
			ret, err := redisPool.Cmd(block.storageModel, key, 1, rowId).Int()
			if ret <= 0 || err != nil {
				log.Errorln("datastorage redis storage error", block.storageModel, key, rowId)
			}
		}
	}
}

//日志消费
func logConsumer(logChannel chan string, pvChannel, uvChannel chan urlData) {
	for logStr := range logChannel {
		data := cutLogFetchData(logStr)

		//uid 模拟
		hasher := md5.New()
		hasher.Write([]byte(data.refer + data.ua))
		uid := hex.EncodeToString(hasher.Sum(nil))

		uData := urlData{
			data,
			uid,
			formatUrl(data.url, data.time),
		}
		pvChannel <- uData
		uvChannel <- uData
	}
}

//格式化日志
func formatUrl(url, t string) urlNode {
	pos1 := str.IndexOf(url, HANDLE_DETAIL, 0)
	if pos1 != -1 {
		pos1 += len(HANDLE_DETAIL)
		pos2 := str.IndexOf(url, HANDLE_HTML, 0)
		idStr := str.Substr(url, pos1, pos2-pos2)
		id, _ := strconv.Atoi(idStr)
		return urlNode{
			"detail",
			id,
			url,
			t,
		}
	} else {
		pos1 = str.IndexOf(url, HANDLE_LIST, 0)
		if pos1 != -1 {
			pos2 := str.IndexOf(url, HANDLE_HTML, 0)
			idStr := str.Substr(url, pos1, pos2-pos2)
			id, _ := strconv.Atoi(idStr)
			return urlNode{
				"list",
				id,
				url,
				t,
			}
		} else {
			return urlNode{
				"home",
				1,
				url,
				t,
			}
		}
	}
}

//逐条格式化日志
func cutLogFetchData(logStr string) digData {
	logStr = strings.TrimSpace(logStr)
	pos1 := str.IndexOf(logStr, HANDLE_DIG, 0)
	if pos1 == -1 {
		return digData{}
	}
	pos1 += len(HANDLE_DIG)
	pos2 := str.IndexOf(logStr, " HTTP/", pos1)
	d := str.Substr(logStr, pos1, pos2-pos1)
	urlInfo, err := url.Parse("http://localhost/?" + d)
	if err != nil {
		return digData{}
	}
	data := urlInfo.Query()
	return digData{
		data.Get("time"),
		data.Get("url"),
		data.Get("refer"),
		data.Get("ua"),
	}
}

//pv统计
func pvCounter(pvChannel chan urlData, storageChannel chan storageBlock) {
	for data := range pvChannel {
		sItem := storageBlock{
			"pv",
			"ZINCRBY",
			data.unode,
		}
		storageChannel <- sItem
	}
}

//uv统计
func uvCounter(uvChannel chan urlData, storageChannel chan storageBlock, redisPool *pool.Pool) {
	for data := range uvChannel {

		hyperLogLogKey := "uv_hpll_" + getTime(data.data.time, "day")
		ret, err := redisPool.Cmd("PFADD", hyperLogLogKey, data.uid, "EX", 86400).Int()
		if err != nil {
			log.Warningln("uvCounter check redis hyperloglog failed " + err.Error())
		}
		if ret != 1 {
			continue
		}
		sItem := storageBlock{
			"uv",
			"ZINCRBY",
			data.unode,
		}
		storageChannel <- sItem
	}
}

func getTime(logTime, timeType string) string {
	var item string
	switch timeType {
	case "day":
		item = "2006-01-02"
		break
	case "hour":
		item = "2006-01-02 15"
		break
	case "min":
		item = "2006-01-02 15:04"
		break
	}
	t, _ := time.Parse(item, logTime)
	return string(strconv.FormatInt(t.Unix(), 10))
}

//逐行读取日志处理
func readFileLinebyLine(params cmdParams, logChannel chan string) error {
	fd, err := os.Open(params.logFilePath)
	if err != nil {
		log.Warning("readFileLinebyLine can't open")
		return err
	}
	defer fd.Close()

	count := 0
	bufferRead := bufio.NewReader(fd)
	for {
		line, err := bufferRead.ReadString('\n')
		logChannel <- line
		count++

		if count%(1000*params.routineNum) == 0 {
			log.Infof("readFileLinebyLine line:%d", count)
		}
		if err != nil {
			if err == io.EOF {
				time.Sleep(3)
				log.Infof("readFileLinebyLine wait, readline:%d", count)
			} else {
				log.Warningf("readFileLinebyLine read err:%s", err.Error())
			}
		}
	}
	return nil
}
