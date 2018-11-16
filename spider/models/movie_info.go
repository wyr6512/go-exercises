package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"regexp"
	"time"
)

var (
	db orm.Ormer
)

const (
	Director = `<a[^>]*rel="v:directedBy">([^<]*)</a>`
	Actors   = `<a[^>]*rel="v:starring">([^<]*)</a>`
	MType    = `<span[^>]*property="v:genre">([^<]*)</span>`
	Name     = `<span[^>]*property="v:itemreviewed">([^<]*)</span>`
	Grade    = `<strong[^>]*property="v:average">([^<]*)</strong>`
	Online   = `<span[^>]*property="v:initialReleaseDate"[^>]*>([^<]*)</span>`
	Image    = `<img[^>]*src="([^>]*)"[^>]*rel="v:image"[^>]*/>`
	Urls     = `<dd>[^<]*<a[^>]*href="([^\"]*)"[^>]*>[^<]*</a>[^<]*</dd>`
)

type MovieInfo struct {
	Id            int64
	MovieId       int64
	MovieName     string
	MoviePic      string
	MovieDirector string
	MovieCountry  string
	MovieLanguage string
	MovieActors   string
	MovieType     string
	MovieOnline   string
	MovieGrade    string
}

func init() {
	//orm.Debug = true
	orm.DefaultTimeLoc = time.UTC
	orm.RegisterDataBase("default", "mysql", "root:@tcp(127.0.0.1:3306)/test?charset=utf8&loc=Local", 30)
	orm.RegisterModel(new(MovieInfo))
	db = orm.NewOrm()
}

func AddMovie(movie_info *MovieInfo) (int64, error) {
	id, err := db.Insert(movie_info)
	return id, err
}

func GetItems(html string, reg string) (data []string) {
	if html == "" {
		return nil
	}
	regex := regexp.MustCompile(reg)
	res := regex.FindAllStringSubmatch(html, -1)
	for _, v := range res {
		data = append(data, v[1])
	}
	return data
}
