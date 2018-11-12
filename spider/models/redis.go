package models

import (
	"github.com/astaxie/goredis"
)

const (
	URL_QUEUE    = "movie_queue"
	URL_USED_SET = "movie_used_set"
)

var (
	client goredis.Client
)

func ConnectRedis(addr string) {
	client.Addr = addr
}

func InQueue(url string) {
	client.Lpush(URL_QUEUE, []byte(url))
}

func OutQueue() string {
	res, err := client.Rpop(URL_QUEUE)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func AddToSet(url string) {
	client.Sadd(URL_USED_SET, []byte(url))
}

func IsVisit(url string) bool {
	b, err := client.Sismember(URL_USED_SET, []byte(url))
	if err != nil {
		return false
	}
	return b
}

func GetQueueLen() int {
	length, err := client.Llen(URL_QUEUE)
	if err != nil {
		return 0
	}
	return length
}
