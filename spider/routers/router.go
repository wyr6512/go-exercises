package routers

import (
	"github.com/astaxie/beego"
	"go-exercises/spider/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/get_movie", &controllers.SpiderController{}, "*:GetMovie")
}
