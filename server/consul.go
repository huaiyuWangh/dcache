package server

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"net/http"
)

var client *consulapi.Client
var serverName = "dcache"

func Register(port int) {
	config := consulapi.DefaultConfig()
	config.Address = ":8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = fmt.Sprintf("%s_%s_%d", serverName, localIP, port) // 服务节点的名称
	registration.Name = serverName                                            // 服务名称
	registration.Port = port                                                  // 服务端口
	registration.Tags = []string{"v1"}                                        // tag，可以为空
	registration.Address = localIP                                       // 服务 IP
	checkPort := port+1000
	registration.Check = &consulapi.AgentServiceCheck{ // 健康检查
		HTTP:                           fmt.Sprintf("http://%s:%d%s", registration.Address, checkPort, "/check"),
		Timeout:                        "3s",
		Interval:                       "5s",  // 健康检查间隔
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务，注销时间，相当于过期时间
		// GRPC:     fmt.Sprintf("%v:%v/%v", IP, r.Port, r.Service),// grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal("register server error : ", err)
	}

	http.HandleFunc("/check", consulCheck)
	http.ListenAndServe(fmt.Sprintf(":%d", checkPort), nil)
}

// consul 服务端会自己发送请求，来进行健康检查
func consulCheck(w http.ResponseWriter, r *http.Request) {
	s := "consulCheck " + " remote:" + r.RemoteAddr + " " + r.URL.String()
	fmt.Fprintln(w, s)
}

func init() {
	config := consulapi.DefaultConfig()
	config.Address = ":8500" //consul server
	var err error
	client, err = consulapi.NewClient(config)
	if err != nil {
		fmt.Println("api new client is failed, err:", err)
		return
	}

}
func Discover(peers *HTTPPool) {
	var lastIndex uint64
	for {
		services, metainfo, err := client.Health().Service(serverName, "v1", true, &consulapi.QueryOptions{
			WaitIndex: lastIndex,
		})
		if err != nil {
			log.Printf("error retrieving instances from Consul: %v\n", err)
		}
		lastIndex = metainfo.LastIndex
		addrs := make([]string, 0)
		for _, service := range services {
			ip := service.Service.Address
			port := service.Service.Port
			fmt.Println("service.Service.Address:", ip, "service.Service.Port:", port)
			addrs = append(addrs, fmt.Sprintf("%s:%d",ip,port))
		}
		peers.Set(addrs...)
	}
}
