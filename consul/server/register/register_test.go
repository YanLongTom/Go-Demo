package register

import (
	cloud "Demo/consul/server/serverInform"
	util "Demo/consul/utils"
	"github.com/gin-gonic/gin"
	"testing"
)

func TestConsulServiceRegistry(t *testing.T) {
	host := "127.0.0.1"
	port := 8500
	registryDiscoveryClient, _ := NewConsulServiceRegistry(host, port, "")

	_, err := util.GetLocalIP()
	if err != nil {
		t.Error(err)
	}

	serviceInstanceInfo, _ := cloud.NewDefaultServiceInstance("go-user-server", "", 8090,
		false, map[string]string{"user": "zyn"}, "")

	success := registryDiscoveryClient.Register(serviceInstanceInfo)
	if !success {
		t.Error("Register failed")
	}

	r := gin.Default()
	// 健康检测接口，其实只要是 200 就认为成功了
	r.GET("/actuator/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	err = r.Run(":8090")
	if err != nil {
		registryDiscoveryClient.Deregister()
	}
}
