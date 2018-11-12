package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"go-exercises/spider/models"
	"strings"
	"time"
)

type SpiderController struct {
	beego.Controller
}

func (c *SpiderController) GetMovie() {
	var movieInfo models.MovieInfo
	models.ConnectRedis("127.0.0.1:6379")              //连接redis
	url := "https://movie.douban.com/subject/3168101/" //指定抓取入口页面
	models.InQueue(url)                                //url入队列
	for {
		length := models.GetQueueLen() //队列长度
		if length == 0 {               //队列为空结束
			break
		}
		url = models.OutQueue()  //出队列
		if models.IsVisit(url) { //已抓取，跳过
			continue
		}
		rsp := httplib.Get(url)
		models.AddToSet(url) //添加到抓取集合
		html, err := rsp.String()
		if err != nil {
			continue
		}
		names := models.GetItems(html, models.Name)
		if names != nil && names[0] != "" { //电影名为空
			movieInfo.MovieName = names[0]
			movieInfo.MovieActors = strings.Join(models.GetItems(html, models.Actors), "/")
			movieInfo.MovieDirector = strings.Join(models.GetItems(html, models.Director), "/")
			movieInfo.MovieGrade = strings.Join(models.GetItems(html, models.Grade), "/")
			movieInfo.MovieOnline = strings.Join(models.GetItems(html, models.Online), "/")
			movieInfo.MoviePic = strings.Join(models.GetItems(html, models.Image), "/")
			_, err = models.AddMovie(&movieInfo) //入库
			if err != nil {
				continue
			}
		}
		urls := models.GetItems(html, models.Urls) //抓取页面url
		for _, u := range urls {
			if !models.IsVisit(u) { //未抓取url入队列
				models.InQueue(u)
			}
		}
		time.Sleep(time.Second * 1)
	}
}
