# simplek8s

> simple k8s implement which inspired by original k8s
> 最简 k8s 实现，由 k8s 的第一个 commit 改造而来

[README中文](./README_ZH.md)

## Abstract

k8s is the current de facto standard in the cloud-native field.   
Thanks to Google's ingenious design, k8s can meet almost all cloud-native scenarios, but it brings a huge threshold for
beginners.

Based on the minimal system learning methodology, we start from the very beginning.   
Commit starts to see how one of the most sophisticated k8s is designed and implemented.

The implementation of k8s is exquisite and complex, and it is a master in the field of golang.   
Learning it well is not only very helpful for work,   
but also can better understand the current business scenarios such as cloud computing and cloud native.

I hope everyone who comes into contact with k8s can gain a lot.

This is my [study notes](./docs/notes.md) in Chinese.

## Architecture

![architecture](./docs/simplek8s.png)

## setup

```sh
./scripts/etcd2.sh
./third_party/etcd-download-test/etcd
make build 
./cmd/integration/integration
```

you will see:

```
2022/09/29 18:00:58 Creating etcd client pointing to [http://localhost:4001]
2022/09/29 18:00:58 API Server started on http://127.0.0.1:45599
2022/09/29 18:00:58 Synchronizing php
2022/09/29 18:00:58 GET /api/v1beta1/tasks
.....
2022/09/29 18:01:08 Too many replicas, deleting
2022/09/29 18:01:08 DELETE /api/v1beta1/tasks/7aaadc5700bf1d9a
2022/09/29 18:01:08 DELETE /api/v1beta1/tasks/53663980e406eccf
2022/09/29 18:01:08 Synchronizing nginxController
2022/09/29 18:01:08 GET /api/v1beta1/tasks?labels=name%3Dnginx
2022/09/29 18:01:08 []api.Task(nil)
2022/09/29 18:01:08 Too few replicas, creating 2
2022/09/29 18:01:08 POST /api/v1beta1/tasks
2022/09/29 18:01:08 POST /api/v1beta1/tasks
2022/09/29 18:01:18 GET /api/v1beta1/tasks
2022/09/29 18:01:18 OK
```

## references

- [kubernetes](https://github.com/kubernetes/kubernetes)
- [learning k8s by prs and issues](https://github.com/cit965/k8s)