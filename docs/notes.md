# 笔记

## first commit

切到 [first commit](https://github.com/kubernetes/kubernetes/commit/2c4b3a562ce34cddc3f8218a2c4d11c7310e6d56)：

```sh
$ git log --reverse

$ git checkout 2c4b3a562ce34cddc3f8218a2c4d11c7310e6d56
$ git switch -c first-commit
$ git clean -d -f
```

## etcd

暂时使用老版本的 etcd，后续会迁移到最新的 etcd 上。

打开 etcd 老版本的[链接](https://github.com/etcd-io/etcd/releases/tag/v2.0.9)，下载 etcd 2.0.9，如下：

```sh
curl -L  https://github.com/coreos/etcd/releases/download/v2.0.9/etcd-v2.0.9-linux-amd64.tar.gz -o etcd-v2.0.9-linux-amd64.tar.gz
tar xzvf etcd-v2.0.9-linux-amd64.tar.gz
cd etcd-v2.0.9-linux-amd64
./etcd
```

打开另一个窗口：

```sh
./etcdctl set mykey "this is awesome"
./etcdctl get mykey
```

> 提示：新版本的 MacOS 运行 2.0.9 版本的 etcd 会报错，推荐在文档的 centos 上来运行，或者直接使用 Docker。

然后运行 make 得到所有的可执行文件：

```sh
make

# 运行控制管理器(调度器)
./cmd/controller-manager/controller-manager -etcd_servers http://localhost:4001 -master http://localhost:8080

# 运行 apiserver
./cmd/apiserver/apiserver -etcd_servers http://localhost:4001 -machines 127.0.0.1

# 运行容器管理器，创建、运行、销毁容器
./cmd/kubelet/kubelet -etcd_servers http://localhost:4001

# 运行代理服务
./cmd/proxy/proxy -etcd_servers http://localhost:4001

# 运行 api 控制客户端
./cmd/kubectl/kubectl -h http://127.0.0.1:8080 -c ./frontend.json create /task
./cmd/kubectl/kubectl -h http://127.0.0.1:8080 get /tasks
```

## 概念

- Manifest：表现
- apiserver：服务
- kubectl：cli，操作 apiserver，后面更名为 kubectl
- controller-manager：管理器，调度器，比如容器复制、备份等
- kubelet：操作 docker 等容器
- proxy：代理服务
- task：容器任务，后面更名为 pod
- api/v1beta1：其实就是 api url 的前缀
- manifest：对资源数据的一种抽象

## 关键点

- 服务注册可以基于 etcd 或者 memory