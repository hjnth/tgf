package internal

import (
	"fmt"
	"github.com/cornelk/hashmap"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv/store"
	"github.com/rpcxio/rpcx-consul/client"
	"github.com/rpcxio/rpcx-consul/serverplugin"
	client2 "github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/thkhxm/tgf"
	"github.com/thkhxm/tgf/log"
	"github.com/thkhxm/tgf/util"
	"time"
)

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ 277949041
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/2/23
//***************************************************

// TODO 修改consul的配置，调整心跳间隔

type ConsulDiscovery struct {
	discoveryMap *hashmap.Map[string, *client.ConsulDiscovery]
}

func (this *ConsulDiscovery) initStruct() {
	var ()
	this.discoveryMap = hashmap.New[string, *client.ConsulDiscovery]()
}

func (this *ConsulDiscovery) RegisterServer(ip string) server.Plugin {
	var (
		address        = tgf.GetStrListConfig(tgf.EnvironmentConsulAddress)
		serviceAddress = fmt.Sprintf("tcp@%v", ip)
		_logAddressMsg string
		_basePath      = tgf.GetStrConfig[string](tgf.EnvironmentConsulPath)
	)
	//注册服务发现根目录
	r := &serverplugin.ConsulRegisterPlugin{
		ServiceAddress: serviceAddress,
		ConsulServers:  address,
		BasePath:       _basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Second * 11,
	}
	err := r.Start()
	if err != nil {
		log.Error("[init] 服务发现启动异常 %v", err)
	}
	for _, s := range address {
		_logAddressMsg += s + ","
	}
	log.Info("[init] 服务发现加载成功 注册根目录 consulAddress=%v serviceAddress=%v path=%v", r.ServiceAddress, _logAddressMsg, _basePath)
	return r
}

func (this *ConsulDiscovery) RegisterClient(serviceName string) client2.XClient {
	var (
		option = client2.DefaultOption
	)
	//if this.configService.IsGateway() {
	//	option.SerializeType = protocol.SerializeNone
	//}
	discovery := this.GetDiscovery(serviceName)
	client := client2.NewXClient(serviceName, client2.Failover, client2.SelectByUser, discovery, option)
	return client
}

func (this *ConsulDiscovery) GetDiscovery(moduleName string) *client.ConsulDiscovery {
	if val, ok := this.discoveryMap.Get(moduleName); ok {
		return val
	}
	var (
		address  = tgf.GetStrListConfig(tgf.EnvironmentConsulAddress)
		basePath = tgf.GetStrConfig[string](tgf.EnvironmentConsulPath)
	)

	//new discovery

	conf := &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: 0,
		Bucket:            "",
		PersistConnection: false,
		Username:          "",
		Password:          "",
	}
	d, _ := client.NewConsulDiscovery(basePath, moduleName, address, conf)
	this.discoveryMap.Set(moduleName, d)
	util.Go(func() {
		for {
			select {
			case kv := <-d.WatchService():
				for _, v := range kv {
					log.Debug("[consul] watch %v service %v,%v", moduleName, v.Key, v.Value)
				}
			}
		}
	})
	return d
}
