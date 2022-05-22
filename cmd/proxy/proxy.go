/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"flag"
	"log"
	"os"

	"github.com/coreos/go-etcd/etcd"

	"github.com/pedrogao/simplek8s/pkg/proxy"
	"github.com/pedrogao/simplek8s/pkg/proxy/config"
)

var (
	configFile  = flag.String("configfile", "/tmp/proxy_config", "Configuration file for the proxy")
	etcdServers = flag.String("etcd_servers", "http://10.240.10.57:4001", "Servers for the etcd cluster (http://ip:port).")
)

func main() {
	// 代理：
	// 分别从 etcd，文件中读取配置，然后负载均衡
	flag.Parse()

	// Set up logger for etcd client
	etcd.SetLogger(log.New(os.Stderr, "etcd ", log.LstdFlags))

	log.Printf("Using configuration file %s and etcd_servers %s", *configFile, *etcdServers)

	proxyConfig := config.NewServiceConfig()

	// Create a configuration source that handles configuration from etcd.
	etcdClient := etcd.NewClient([]string{*etcdServers})
	config.NewConfigSourceEtcd(etcdClient,
		proxyConfig.GetServiceConfigurationChannel("etcd"),
		proxyConfig.GetEndpointsConfigurationChannel("etcd"))

	// And create a configuration source that reads from a local file
	config.NewConfigSourceFile(*configFile,
		proxyConfig.GetServiceConfigurationChannel("file"),
		proxyConfig.GetEndpointsConfigurationChannel("file"))

	loadBalancer := proxy.NewLoadBalancerRR()
	proxier := proxy.NewProxier(loadBalancer)
	// Wire proxier to handle changes to services
	proxyConfig.RegisterServiceHandler(proxier)
	// And wire loadBalancer to handle changes to endpoints to services
	proxyConfig.RegisterEndpointsHandler(loadBalancer)

	// Just loop forever for now...
	select {}

}
