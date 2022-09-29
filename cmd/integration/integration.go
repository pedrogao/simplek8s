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
// A basic integration test for the service.
// Assumes that there is a pre-existing etcd server running on localhost.
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"time"

	"github.com/coreos/go-etcd/etcd"

	"github.com/pedrogao/simplek8s/pkg/api"
	"github.com/pedrogao/simplek8s/pkg/apiserver"
	kubeclient "github.com/pedrogao/simplek8s/pkg/client"
	"github.com/pedrogao/simplek8s/pkg/registry"
)

func main() {
	// 本地集成测试
	// Setup
	servers := []string{"http://localhost:4001"}
	log.Printf("Creating etcd client pointing to %v", servers)
	etcdClient := etcd.NewClient(servers)
	machineList := []string{"machine"}
	// ectd 注册组件
	reg := registry.MakeEtcdRegistry(etcdClient, machineList)
	// apiserver
	apiserver := apiserver.New(
		map[string]apiserver.RESTStorage{
			"tasks": registry.MakeTaskRegistryStorage(reg, &kubeclient.FakeContainerInfo{},
				registry.MakeRoundRobinScheduler(machineList)), // 任务，后面改为 pod
			"replicationControllers": registry.MakeControllerRegistryStorage(reg), // 控制器
		},
		"/api/v1beta1", // URL 路径
	)
	server := httptest.NewServer(apiserver)
	// 控制管理器
	controllerManager := registry.MakeReplicationManager(
		etcd.NewClient(servers),
		kubeclient.Client{
			Host: server.URL,
		},
	)
	go controllerManager.Synchronize()      // 同步元数据
	go controllerManager.WatchControllers() // 监听控制器

	// Ok. we're good to go.
	log.Printf("API Server started on %s", server.URL)
	// Wait for the synchronization threads to come up.
	time.Sleep(time.Second * 10)

	// 开启客户端请求是否成功
	kubeClient := kubeclient.Client{
		Host: server.URL,
	}
	data, err := ioutil.ReadFile("docs/controller.json")
	if err != nil {
		log.Fatalf("Unexpected error: %#v", err)
	}
	var controllerRequest api.ReplicationController
	if err = json.Unmarshal(data, &controllerRequest); err != nil {
		log.Fatalf("Unexpected error: %#v", err)
	}

	if _, err = kubeClient.CreateReplicationController(controllerRequest); err != nil {
		log.Fatalf("Unexpected error: %#v", err)
	}
	// Give the controllers some time to actually create the tasks
	time.Sleep(time.Second * 10)

	// Validate that they're truly up.
	tasks, err := kubeClient.ListTasks(nil)
	if err != nil || len(tasks.Items) != 2 {
		log.Fatal("FAILED")
	}
	log.Printf("OK")
}
