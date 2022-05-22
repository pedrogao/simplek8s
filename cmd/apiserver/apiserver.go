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
// apiserver is the main api server and master for the cluster.
// it is responsible for serving the cluster management API.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-etcd/etcd"

	"github.com/pedrogao/simplek8s/pkg/apiserver"
	kubeclient "github.com/pedrogao/simplek8s/pkg/client"
	"github.com/pedrogao/simplek8s/pkg/registry"
	"github.com/pedrogao/simplek8s/pkg/util"
)

var (
	port                        = flag.Uint("port", 8080, "The port to listen on.  Default 8080.")
	address                     = flag.String("address", "127.0.0.1", "The address on the local server to listen to. Default 127.0.0.1")
	apiPrefix                   = flag.String("api_prefix", "/api/v1beta1", "The prefix for API requests on the server. Default '/api/v1beta1'")
	etcdServerList, machineList util.StringList
)

func init() {
	flag.Var(&etcdServerList, "etcd_servers", "Servers for the etcd (http://ip:port), comma separated")
	flag.Var(&machineList, "machines", "List of machines to schedule onto, comma separated.")
}

func main() {
	flag.Parse()

	if len(machineList) == 0 {
		log.Fatal("No machines specified!")
	}

	var (
		taskRegistry       registry.TaskRegistry       // pod 注册
		controllerRegistry registry.ControllerRegistry // 控制器注册
		serviceRegistry    registry.ServiceRegistry    // 服务注册
	)
	if len(etcdServerList) > 0 {
		// 基于 etcd
		log.Printf("Creating etcd client pointing to %v", etcdServerList)
		etcdClient := etcd.NewClient(etcdServerList)
		taskRegistry = registry.MakeEtcdRegistry(etcdClient, machineList)
		controllerRegistry = registry.MakeEtcdRegistry(etcdClient, machineList)
		serviceRegistry = registry.MakeEtcdRegistry(etcdClient, machineList)
	} else {
		// 基于内存
		taskRegistry = registry.MakeMemoryRegistry()
		controllerRegistry = registry.MakeMemoryRegistry()
		serviceRegistry = registry.MakeMemoryRegistry()
	}
	// 容器客户端
	containerInfo := &kubeclient.HTTPContainerInfo{
		Client: http.DefaultClient,
		Port:   10250,
	}

	storage := map[string]apiserver.RESTStorage{
		// 任务存储多了一个调度器，负责调度任务，将选择合适的机器去执行任务
		"tasks": registry.MakeTaskRegistryStorage(taskRegistry, containerInfo,
			registry.MakeFirstFitScheduler(machineList, taskRegistry)),                       // 任务存储
		"replicationControllers": registry.MakeControllerRegistryStorage(controllerRegistry), // 控制器存储
		"services":               registry.MakeServiceRegistryStorage(serviceRegistry),       // 服务存储
	}
	// 服务、任务
	endpoints := registry.MakeEndpointController(serviceRegistry, taskRegistry)

	// 端点元数据，如：某个服务在某个机器下的某个端口
	go util.Forever(func() { endpoints.SyncServiceEndpoints() }, time.Second*10) // 每 10s 刷新一次

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", *address, *port),
		Handler:        apiserver.New(storage, *apiPrefix),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1M
	}
	log.Fatal(s.ListenAndServe())
}
