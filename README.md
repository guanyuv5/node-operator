## node-operator demo with operator sdk
### 环境准备

```bash
# wget https://golang.org/dl/go1.13.15.linux-amd64.tar.gz
# tar xvf go1.13.15.linux-amd64.tar.gz
export GOROOT=/root/go
export GOPATH=/root/gopath
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash

# export RELEASE_VERSION=v0.19.0
# curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
# chmod +x operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu && sudo mkdir -p /usr/local/bin/ && sudo cp operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/operator-sdk && rm operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
```

Operator SDK 提供以下工作流来开发新的 Operator：  

```text
1. 使用 Operator SDK 命令行界面 (CLI) 新建一个 Operator 项目。
2. 通过添加自定义资源定义 (CRD) 来定义新的资源 API。
3. 在指定的处理程序中定义 Operator 协调逻辑，并使用 Operator SDK API 与资源交互。
4. 使用 Operator SDK CLI 来构建和生成 Operator 部署组件。
```

### node-operator demo开发
#### 1. 创建node-operator项目

```
# mkdir node-operator && cd node-operator/
# operator-sdk init --project-version=2 --license apache2 --domain=jike.com --repo=github.com/jike-inc/node-operator
Writing scaffold for you to edit...
Get controller runtime:
$ go get sigs.k8s.io/controller-runtime@v0.6.2
Update go.mod:
$ go mod tidy
Running make:
$ make
/Users/xxx/code/gopath/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
Next: define a resource with:
$ operator-sdk create api

~/workspace/node-operator  tree -L 2
.
├── Dockerfile
├── Makefile
├── PROJECT
├── bin
│   └── manager
├── config
│   ├── certmanager
│   ├── default
│   ├── manager
│   ├── prometheus
│   ├── rbac
│   ├── scorecard
│   └── webhook
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt
└── main.go
```

#### 2. 声明 api 和控制器模板：

```bash
# operator-sdk create api --group=nodes --version=v1alpha1 --kind=NodeOP
Create Resource [y/n]
y
Create Controller [y/n]
y
Writing scaffold for you to edit...
api/v1alpha1/nodeop_types.go
controllers/nodeop_controller.go
Running make:
$ make
/Users/guanyuding/code/gopath/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go build -o bin/manager main.go
```

#### 3. node-oprator 实现逻辑
- 3.1 增加自定义SPEC

修改api/v1alpha1/nodeop_types.go文件中的NodeOPSpec结构体:

```golang
// NodeOPSpec defines the desired state of NodeOP
type NodeOPSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
```

- 3.2 实现Reconcile接口，完善业务逻辑

```golang
func (r *NodeOPReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	...
}
```

具体查看controllers/nodeop_controller.go 代码


#### 4. 使用 Operator SDK CLI 来构建和生成 Operator

- 自动生成代码

```bash
# make generate
```
- 生成crd资源

```bash
$ make manifests
config/crd/bases/cache.example.com_memcacheds.yaml
```
- 编译operator：

```bash
# GOOS=linux GOARCH=amd64 make
/Users/xxx/code/gopath/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
main.go
go vet ./...
go build -o bin/manager main.go
```

### 测试

- 向k8s集群中创建crd:

```bash
# kubectl create -f config/crd/bases/nodes.jike.com_nodeops.yaml
customresourcedefinition.apiextensions.k8s.io/nodeops.nodes.jike.com created
```

- 启动manager：

```bash
# ./manager --kubeconfig=/root/.kube/config --hostname=172.21.240.97 --metrics-addr=0.0.0.0:8090
2020-09-02T22:20:21.857+0800	INFO	controller-runtime.metrics	metrics server is starting to listen	{"addr": ":8080"}
2020-09-02T22:20:21.858+0800	INFO	setup	starting manager
2020-09-02T22:20:21.858+0800	INFO	controller-runtime.manager	starting metrics server	{"path": "/metrics"}
2020-09-02T22:20:21.858+0800	INFO	controller	Starting EventSource	{"reconcilerGroup": "nodes.jike.com", "reconcilerKind": "NodeOP", "controller": "nodeop", "source": "kind source: /, Kind="}
2020-09-02T22:20:21.958+0800	INFO	controller	Starting Controller	{"reconcilerGroup": "nodes.jike.com", "reconcilerKind": "NodeOP", "controller": "nodeop"}
2020-09-02T22:20:21.958+0800	INFO	controller	Starting workers	{"reconcilerGroup": "nodes.jike.com", "reconcilerKind": "NodeOP", "controller": "nodeop", "worker count": 1}
参数说明：
--kubeconfig： 指定连接kube-apiserver的地址信息
--hostname: 指定operator所在的节点ip，类似于kubelet中的--host-overwrite参数
```

- 定义cr并创建：

```yaml
# cat cr.yaml
apiVersion: nodes.jike.com/v1alpha1
kind: NodeOP
metadata:
  name: 172.21.240.97
spec:
  command: ifconfig
  args: ["eth0"]
  
# kubectl create -f cr.yaml
nodeop.nodes.jike.com/172.21.240.97 created
```

- 查看 manager 日志输出:

```text
2020-09-02T22:22:41.701+0800	INFO	controllers.NodeOP	ifconfig,[eth0]	{"memcached": "default/172.21.240.97"}
combined out:
eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 172.21.240.97  netmask 255.255.255.0  broadcast 172.21.240.255
        inet6 fe80::5054:ff:fe65:d6d7  prefixlen 64  scopeid 0x20<link>
        ether 52:54:00:65:d6:d7  txqueuelen 1000  (Ethernet)
        RX packets 13337896  bytes 4837308835 (4.5 GiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 11897978  bytes 3136867279 (2.9 GiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```

### 参照
[https://sdk.operatorframework.io/docs/building-operators/golang/]()  
[https://github.com/operator-framework/operator-sdk]()
