package panel

import (
	"sun-panel/api/api_v1"
	"sun-panel/api/api_v1/middleware"

	"github.com/gin-gonic/gin"
)

func InitNetwork(router *gin.RouterGroup) {
	api := api_v1.ApiGroupApp.ApiPanel.NetworkApi
	r := router.Group("", middleware.LoginInterceptor, middleware.AdminInterceptor)
	{
		r.GET("/panel/network/probe", api.GetProbeConfig)
		r.POST("/panel/network/probe", api.SaveProbeConfig)
		r.POST("/panel/network/probe/test", api.TestProbe)
	}
}
