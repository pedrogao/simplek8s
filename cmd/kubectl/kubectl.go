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
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	kubeclient "github.com/pedrogao/simplek8s/pkg/client"
	"github.com/pedrogao/simplek8s/pkg/kubectl"
)

const AppVersion = "0.1"

// The flag package provides a default help printer via -h switch
var (
	versionFlag  = flag.Bool("v", false, "Print the version number.")
	httpServer   = flag.String("h", "", "The host to connect to.")
	config       = flag.String("c", "", "Path to the config file.")
	labelQuery   = flag.String("l", "", "Label query to use for listing")
	updatePeriod = flag.Duration("u", 60*time.Second, "Update interarrival in seconds")
	portSpec     = flag.String("p", "", "The port spec, comma-separated list of <external>:<internal>,...")
	servicePort  = flag.Int("s", -1, "If positive, create and run a corresponding service on this port, only used with 'run'")
	authConfig   = flag.String("auth", os.Getenv("HOME")+"/.kubernetes_auth", "Path to the auth info file.  If missing, prompt the user")
)

func usage() {
	log.Fatal("Usage: kubectl -h <host> [-c config/file.json] [-p <hostPort>:<containerPort>,..., <hostPort-n>:<containerPort-n> <method> <path>")
}

// Kubectl command line tool.
func main() {
	flag.Parse() // Scan the arguments list

	if *versionFlag {
		fmt.Println("Version:", AppVersion)
		os.Exit(0)
	}

	if len(flag.Args()) < 2 {
		usage()
	}
	method := flag.Arg(0)
	url := *httpServer + "/api/v1beta1" + flag.Arg(1)
	var request *http.Request
	var err error

	auth, err := kubectl.LoadAuthInfo(*authConfig) // 加载姓名、密码
	if err != nil {
		log.Fatalf("Error loading auth: %#v", err)
	}

	if method == "get" || method == "list" {
		if len(*labelQuery) > 0 && method == "list" {
			url = url + "?labels=" + *labelQuery
		}
		request, err = http.NewRequest("GET", url, nil)
	} else if method == "delete" {
		request, err = http.NewRequest("DELETE", url, nil)
	} else if method == "create" {
		request, err = kubectl.RequestWithBody(*config, url, "POST")
	} else if method == "update" {
		request, err = kubectl.RequestWithBody(*config, url, "PUT")
	} else if method == "rollingupdate" {
		client := &kubeclient.Client{
			Host: *httpServer,
			Auth: &auth,
		}
		kubectl.Update(flag.Arg(1), client, *updatePeriod)
	} else if method == "run" {
		args := flag.Args()
		if len(args) < 4 {
			log.Fatal("usage: kubectl -h <host> run <image> <replicas> <name>")
		}
		image := args[1]
		replicas, err := strconv.Atoi(args[2]) // 备份数量
		name := args[3]
		if err != nil {
			log.Fatalf("Error parsing replicas: %#v", err)
		}
		err = kubectl.RunController(image, name, replicas, kubeclient.Client{Host: *httpServer, Auth: &auth}, *portSpec, *servicePort)
		if err != nil {
			log.Fatalf("Error: %#v", err)
		}
		return
	} else if method == "stop" {
		err = kubectl.StopController(flag.Arg(1), kubeclient.Client{Host: *httpServer, Auth: &auth})
		if err != nil {
			log.Fatalf("Error: %#v", err)
		}
		return
	} else if method == "rm" {
		err = kubectl.DeleteController(flag.Arg(1), kubeclient.Client{Host: *httpServer, Auth: &auth})
		if err != nil {
			log.Fatalf("Error: %#v", err)
		}
		return
	} else {
		log.Fatalf("Unknown command: %s", method)
	}
	if err != nil {
		log.Fatalf("Error: %#v", err)
	}
	var body string
	body, err = kubectl.DoRequest(request, auth.User, auth.Password)
	if err != nil {
		log.Fatalf("Error: %#v", err)
	}
	fmt.Println(body)
}
