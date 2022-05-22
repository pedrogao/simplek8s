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
// The controller manager is responsible for monitoring replication controllers, and creating corresponding
// tasks to achieve the desired state.  It listens for new controllers in etcd, and it sends requests to the
// master to create/delete tasks.
//
// TODO: Refactor the etcd watch code so that it is a pluggable interface.
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/coreos/go-etcd/etcd"

	kubeclient "github.com/pedrogao/simplek8s/pkg/client"
	"github.com/pedrogao/simplek8s/pkg/registry"
	"github.com/pedrogao/simplek8s/pkg/util"
)

var (
	etcdServers = flag.String("etcd_servers", "", "Servers for the etcd (http://ip:port).")
	master      = flag.String("master", "", "The address of the Kubernetes API server")
)

func main() {
	// 检测副本控制器，当新的任务出现时，调用控制器接口使任务到达理想状态
	flag.Parse()

	if len(*etcdServers) == 0 || len(*master) == 0 {
		log.Fatal("usage: controller-manager -etcd_servers <servers> -master <master>")
	}

	// Set up logger for etcd client
	etcd.SetLogger(log.New(os.Stderr, "etcd ", log.LstdFlags))

	controllerManager := registry.MakeReplicationManager(
		etcd.NewClient([]string{*etcdServers}),
		kubeclient.Client{
			Host: "http://" + *master,
		},
	)
	// 感觉有点重复，既不断的请求数据，又 watch 监听
	go util.Forever(func() { controllerManager.Synchronize() }, 20*time.Second)
	go util.Forever(func() { controllerManager.WatchControllers() }, 20*time.Second)

	select {}
}
