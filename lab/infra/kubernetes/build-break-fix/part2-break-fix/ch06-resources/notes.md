# k8sのリソースを作って壊してみよう

## Podのライフサイクルに関して

Pending -> Running -> Succeeded/Failed
RunningがPodが起動している状態

## Podを冗長化するためにのReplicaSetとDeployment

- ReplicaSetはPodの複製を管理するリソース
- DeploymentはReplicaSetを管理するリソースで、ローリングアップデートやロールバックなどの機能を提供する

[ReplicaSetとDeploymentの関係](./image/pod.excalidraw.png)

ReplicaSetのyamlファイルの例
この例では、3のPodが作成されます。

```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: httpserver
  labels:
    app: httpserver
spec:
  replicas: 3
  selector:
    matchLabels:
      app: httpserver
  template:
    metadata:
      labels:
        app: httpserver
    spec:
      containers:
      - name: httpserver
        image: nginx:latest
```

applyして podが3つ作成されるのを確認できます。

```sh
❯ kub get pod
NAME               READY   STATUS              RESTARTS   AGE
httpserver-6mljv   0/1     ContainerCreating   0          12s
httpserver-74dxt   0/1     ContainerCreating   0          12s
httpserver-jdllt   0/1     ContainerCreating   0          12s
```

冗長化を考えると、ReplicaSetのみで十分に感じますが、Deploymentを使用することで、ローリングアップデートやロールバックなどの機能を利用できるため、より柔軟な運用が可能になります。


Deploymentのyamlファイルの例
この例では、1つのDeploymentが作成され、ReplicaSetが3つのPodを管理します。

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:1.24.0
          ports:
            - containerPort: 80
```

applyして podが3つ作成されるのを確認できます。

```sh
❯ kub get deployments
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           17s
```

```sh
❯ kub get pod
NAME                                READY   STATUS    RESTARTS   AGE
nginx-deployment-7f9d555b89-jl7z9   1/1     Running   0          27s
nginx-deployment-7f9d555b89-lqffk   1/1     Running   0          27s
nginx-deployment-7f9d555b89-nzbmh   1/1     Running   0          27s
```

spec.template.spec.containers.imageを変更して、ローリングアップデートを試してみましょう。

```yaml
image: nginx:1.25.3
```

applyして、ローリングアップデートが行われるのを確認できます。
replicasetsを確認すると、古いReplicaSetは0のPodを管理し、新しいReplicaSetが3のPodを管理していることがわかります。

```sh
❯ kub get replicasets
NAME                          DESIRED   CURRENT   READY   AGE
nginx-deployment-767df4cb79   3         3         3       34s
nginx-deployment-7f9d555b89   0         0         0       2m25s
```

正しくimageが更新されていることを確認できます。

```sh
❯ kub get deployment nginx-deployment -o=jsonpath='{.spec.template.spec.containers[0].image}'
nginx:1.25.3%
```

### StrategyTypeについて
Deploymentの更新戦略を指定するためのフィールドで、主に以下の2つがあります。
- RollingUpdate: デフォルトの戦略で、Podを順次更新していく方法。ダウンタイムを最小限に抑えることができます。
- Recreate: すべてのPodを一度に削除してから新しいPodを作成する方法。ダウンタイムが発生しますが、リソースの競合を避けることができます。
RollingUpdateを使用する場合、maxUnavailableとmaxSurgeのフィールドを設定することができます。
- maxUnavailable: 更新中に利用できないPodの最大数を指定します。デフォルトは25%です。
- maxSurge: 更新中に作成されるPodの最大数を指定します。デフォルトは25%です。
これらのフィールドを適切に設定することで、更新中の可用性を確保しながら効率的なローリングアップデートを実現できます。

## Deploymentをつくって壊してみよう

下記のyamlファイルをapplyして、Deploymentを作成してみましょう。
> ./k8s/deployment-hello-server.yml

```sh
❯ kub apply -f ./k8s/deployment-hello-server.yml
deployment.apps/hello-server created
```
Deploymentが作成されたことを確認できます。

```sh
❯ kub get pod
NAME                            READY   STATUS    RESTARTS   AGE
hello-server-864c8d69f7-lhsqc   1/1     Running   0          5s
hello-server-864c8d69f7-p2gv8   1/1     Running   0          5s
hello-server-864c8d69f7-vdzlg   1/1     Running   0          5s
```
podを削除

```sh
❯ kub delete pod hello-server-864c8d69f7-lhsqc
pod "hello-server-864c8d69f7-lhsqc" deleted
```

削除したpodが再度作成されることを確認できます。

```sh
❯ kub get pod
NAME                            READY   STATUS    RESTARTS   AGE
hello-server-864c8d69f7-p2gv8   1/1     Running   0          82s
hello-server-864c8d69f7-v5tkl   1/1     Running   0          3s
hello-server-864c8d69f7-vdzlg   1/1     Running   0          82s
```

では port-forwardしてみましょう。

```sh
❯  kub port-forward deployment/hello-server 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

ブラウザで http://localhost:8080 にアクセスしてみましょう。
Hello, World!と表示されることを確認できます。

次に 下記のyamlファイルをapplyして、Deploymentを更新してみましょう。
> ./k8s/deployment-hello-server-rollingupdate.yml

PodのStatusがImagePullBackOffになっていることを確認できます。

```sh
❯ kub get pod
NAME                            READY   STATUS             RESTARTS   AGE
hello-server-7886f99c58-qhjkw   0/1     ImagePullBackOff   0          19s
hello-server-864c8d69f7-p2gv8   1/1     Running            0          5m13s
hello-server-864c8d69f7-v5tkl   1/1     Running            0          3m54s
hello-server-864c8d69f7-vdzlg   1/1     Running            0          5m13s
```

ImagePullBackOffは、指定されたイメージが見つからない場合や、イメージのプルに失敗した場合に発生するエラーです。
同じように、ブラウザで http://localhost:8080 にアクセスしてみましょう。
Hello, World!と表示されることを確認できます。

RollingUpdateの戦略を使用しているため、古いPodは正常に動作し続け、新しいPodが失敗してもサービスが継続されることがわかります。
デフォルトで、maxUnavailableは25%で、maxSurgeも25%であるため、更新中に利用できないPodの最大数は1で、更新中に作成されるPodの最大数も1になります。

これは新しいPodがUP-TO-DATEになっており、古いPodがまだ利用可能であることを意味します。

```sh
❯ kub get deployments
NAME           READY   UP-TO-DATE   AVAILABLE   AGE
hello-server   3/3     1            3           7m56s
```

ReplicaSetを確認すると、古いReplicaSetは3のPodを管理し、新しいReplicaSetは0のPodを管理していることがわかります。

```sh
❯ kub get replicasets
NAME                      DESIRED   CURRENT   READY   AGE
hello-server-7886f99c58   1         1         0       4m3s
hello-server-864c8d69f7   3         3         3       8m57s
```

前回同じ問題なので、editして、imageをhello-server:1.2.0に変更してみましょう。

```sh
❯ kub edit deployment hello-server
```

```sh
❯ kub get pod,replicasets,deployments
NAME                                READY   STATUS    RESTARTS   AGE
pod/hello-server-78d89df559-25llb   1/1     Running   0          61s
pod/hello-server-78d89df559-f4ntl   1/1     Running   0          62s
pod/hello-server-78d89df559-vkp4q   1/1     Running   0          67s

NAME                                      DESIRED   CURRENT   READY   AGE
replicaset.apps/hello-server-7886f99c58   0         0         0       6m39s
replicaset.apps/hello-server-78d89df559   3         3         3       67s
replicaset.apps/hello-server-864c8d69f7   0         0         0       11m

NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/hello-server   3/3     3            3           11m
```

すべてのRollingUpdateが完了したので、port-forwardの接続が切れているはずです。
```sh
❯  kub port-forward deployment/hello-server 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
Handling connection for 8080
Handling connection for 8080
Handling connection for 8080
Handling connection for 8080
Handling connection for 8080
E0526 03:43:34.191156   75414 portforward.go:424] "Unhandled Error" err="an error occurred forwarding 8080 -> 8080: error forwarding port 8080 to pod 9f17366303eb4d65c1de5a507b9e9e2db2c7e3d8b5194df09cca0d0b04d34359, uid : failed to find sandbox \"9f17366303eb4d65c1de5a507b9e9e2db2c7e3d8b5194df09cca0d0b04d34359\" in store: not found"
error: lost connection to pod
```
再度接続して確認してみましょう！


## Podへのアクセスを助けるService
先ほど、RollingUpdateのimageを更新した際に port-forwardの接続が切れてしまいました。
Serviceを作成して、Podへのアクセスを安定させることができます。

[Serviceのイメージ](./image/service.excalidraw.png)

Serviceのyamlファイルの例

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-server-service
spec:
  selector:
    app: hello-server
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
```

このファイルのみではPodが作成されないため、Deploymentを別にapplyして、Podを作成する必要があります。
```sh
❯ kub apply -f ./k8s/deployment-hello-server.yml
deployment.apps/hello-server created
```

```sh
❯ kub apply -f ./k8s/service.yml
service/hello-server-service created
```

実際にServiceが作成されたことを確認できます。

```sh
❯ kub get service
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
hello-server-service   ClusterIP   10.96.221.180   <none>        8080/TCP   3m31s
kubernetes             ClusterIP   10.96.0.1       <none>        443/TCP    42h
```

port-forwardしてみましょう。

```sh
❯ kub port-forward svc/hello-server-service 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
Handling connection for 8080
Handling connection for 8080
```

### ServiceのTypeについて
- ClusterIP: デフォルトのタイプで、クラスター内でのみアクセス可能なIPアドレスを割り当てます。
- NodePort: クラスター外からアクセス可能なポートを割り当てます。NodeのIPアドレスと組み合わせてアクセスできます。
- LoadBalancer: クラウドプロバイダーのロードバランサーを利用して、外部からアクセス可能なIPアドレスを割り当てます。
- ExternalName: DNS名を使用して外部サービスにアクセスするためのタイプです。クラスター内のサービスに対して、外部のDNS名を割り当てます。

一旦新しい Podを作成しても、Serviceを通じてアクセスできることを確認できます。

#### ClusterIP

```sh
❯ kub run curl --image curlimages/curl --rm --stdin --tty --restart=Never --command -- curl 10.96.221.180:8080
Hello, world!pod "curl" deleted
```

別のPodからhello-server-serviceにアクセスできました。
一旦掃除
```sh
❯ kub delete -f k8s/deployment-hello-server.yml
deployment.apps "hello-server" deleted

❯ kub delete -f k8s/service.yml
service "hello-server-service" deleted
```

#### NodePort

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-server-service
spec:
  type: NodePort
  selector:
    app: hello-server
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
      nodePort: 30080
```

```sh
❯ kub apply -f ./k8s/service-nodeport.yml
service/hello-server-service created
```

NodeのIPアドレスを確認して、ブラウザで http://<NodeのIPアドレス>:30080 にアクセスしてみましょう。
Hello, World!と表示されることを確認できます。

Docker Desktopの場合は、localhost:30080でアクセスできます。

#### DNS

先ほど作成した、hello-server-serviceは、クラスター内でhello-server-serviceというDNS名でアクセスできます。

```sh
❯ kub run curl --image curlimages/curl --rm --stdin --tty --restart=Never --command -- curl hello-server-service.default.svc.cluster.local:8080
Hello, world!pod "curl" deleted
```

### Serviceを壊してみる

上で対応したymlをapplyして、Serviceを作成してみましょう。
```sh
❯ kub apply -f ./k8s/service-node-port.yml
service/hello-server-service created
```

NodeのIPアドレスを取得してみよう。

```sh
❯ kub get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'
172.31.0.2
```
今回はkindを使用しているので、localhost:30080でアクセスできます。

```sh
❯ curl localhost:30080
Hello, world!
```

次に、service-destruction.ymlをapplyして、Serviceを壊してみましょう。
```sh
❯ kub apply -f ./k8s/service-destruction.yml
service/hello-server-service configured
```
curlしてみましょう。
```sh
❯ curl localhost:30599
curl: (52) Empty reply from server
```

まずPodから確認
```sh
❯ kub get pod
NAME                            READY   STATUS    RESTARTS   AGE
hello-server-6cc6b44795-2lxch   1/1     Running   0          6m40s
hello-server-6cc6b44795-hvp7j   1/1     Running   0          6m40s
hello-server-6cc6b44795-pj9kh   1/1     Running   0          6m40s
```

次はDeployment

```sh
❯ kub get deployment
NAME           READY   UP-TO-DATE   AVAILABLE   AGE
hello-server   3/3     3            3           7m17s
```

次にServiceを確認してみましょう。

```sh
❯ kub get service
NAME                    TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
hello-server-external   NodePort    10.96.25.51   <none>        8080:30599/TCP   8m16s
kubernetes              ClusterIP   10.96.0.1     <none>        443/TCP          3d5h
```

問題はなさそうですね。
ということは、各何処かのサービスの通信が壊れている可能性があります。
1. Pod 内から アプリケーションがリクエストを受け取る部分
2. クラスタ内のネットワーク
3. Serviceの設定

Podにアクセスして、curlを実行してみましょう。

```sh
❯ kub debug --stdin --tty hello-server-6cc6b44795-2lxch --image curlimages/curl --target=hello-server -- sh

Targeting container "hello-server". If you don't see processes from this container it may be because the container runtime doesn't support this feature.
--profile=legacy is deprecated and will be removed in the future. It is recommended to explicitly specify a profile, for example "--profile=general".
Defaulting debug container name to debugger-4dnrr.
If you don't see a command prompt, try pressing enter.

~ $ curl localhost:8080
Hello, world!~
$
```

問題はないので次の検証です。

同じクラスタで、新しく別のPodを作成してPodにアクセスして外からアクセスできるのかを確認

```sh
❯ kub run curl --image curlimages/curl --rm --stdin --tty --restart=Never --command -- curl 10.244.0.13:8080
Hello, world!pod "curl" deleted
```

次はservice経由で接続確認してみましょう

IPの情報を知らなくても、hello-server-externalというDNS名でアクセスできるはずです。

```sh
❯ kub run curl --image curlimages/curl --rm --stdin --tty --restart=Never --command -- curl hello-server-external:8080
curl: (7) Failed to connect to hello-server-external port 8080 after 8 ms: Could not connect to server
pod "curl" deleted
pod default/curl terminated (Error)
```

connect to serverのエラーが出ているので、Serviceの設定に問題がある可能性があります。

```sh
❯ kub describe services
Name:                     hello-server-external
Namespace:                default
Labels:                   <none>
Annotations:              <none>
Selector:                 app=hello-serve
Type:                     NodePort
IP Family Policy:         SingleStack
IP Families:              IPv4
IP:                       10.96.25.51
IPs:                      10.96.25.51
Port:                     <unset>  8080/TCP
TargetPort:               8080/TCP
NodePort:                 <unset>  30599/TCP
Endpoints:
Session Affinity:         None
External Traffic Policy:  Cluster
Internal Traffic Policy:  Cluster
Events:                   <none>


Name:                     kubernetes
Namespace:                default
Labels:                   component=apiserver
                          provider=kubernetes
Annotations:              <none>
Selector:                 <none>
Type:                     ClusterIP
IP Family Policy:         SingleStack
IP Families:              IPv4
IP:                       10.96.0.1
IPs:                      10.96.0.1
Port:                     https  443/TCP
TargetPort:               6443/TCP
Endpoints:                172.31.0.2:6443
Session Affinity:         None
Internal Traffic Policy:  Cluster
Events:                   <none>
```

ServiceのSelectorに誤りがあることがわかります。
app=hello-serveとなっていますが、正しくはapp=hello-serverです。

diffで見てみましょう

```sh
❯ kub diff -f k8s/service-node-port.yml
diff -u -N /var/folders/kr/z15pp29s0zj3m_17f2sddzhr0000gn/T/LIVE-3153552221/v1.Service.default.hello-server-external /var/folders/kr/z15pp29s0zj3m_17f2sddzhr0000gn/T/MERGED-410664411/v1.Service.default.hello-server-external
--- /var/folders/kr/z15pp29s0zj3m_17f2sddzhr0000gn/T/LIVE-3153552221/v1.Service.default.hello-server-external	2026-05-30 01:29:36
+++ /var/folders/kr/z15pp29s0zj3m_17f2sddzhr0000gn/T/MERGED-410664411/v1.Service.default.hello-server-external	2026-05-30 01:29:36
@@ -24,7 +24,7 @@
     protocol: TCP
     targetPort: 8080
   selector:
-    app: hello-serve
+    app: hello-server
   sessionAffinity: None
   type: NodePort
 status:
```

## ConfigMapを作成してみよう

- コンテナ内のコマンドの引数として読み込む
- コンテナの環境変数として読み込む
- ボリュームとしてマウントして読み込む

#### 環境変数として読み込む

k8s/configmap/hello-server-env.yml に Port 8081でHello, World!を返すサーバーのコードが書いてあります。
このコードをConfigMapに保存して、Podから読み込んでみましょう。

```sh
❯ kub apply -f k8s/configmap/hello-server-env.yml
deployment.apps/hello-server created
configmap/hello-server-configmap created
```

```sh
❯ kub get deployments,configmaps
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/hello-server   1/1     1            1           27s

NAME                               DATA   AGE
configmap/hello-server-configmap   1      27s
configmap/kube-root-ca.crt         1      3m10s
```

port-forwardしてみましょう。

```sh
❯ kub port-forward deployments/hello-server 8081:8081
Forwarding from 127.0.0.1:8081 -> 8081
Forwarding from [::1]:8081 -> 8081
```

curlしてみましょう。

```sh
❯ curl localhost:8081
Hello, world! Let's learn Kubernetes!%
```

#### ボリュームとしてマウントして読み込む

k8s/configmap/hello-server-volume.yml に myconfig.txt というファイルをマウントして、ファイルの内容を返すサーバーのコードが書いてあります。
このコードをConfigMapに保存して、Podから読み込んでみましょう。

```sh
❯ kub apply -f k8s/configmap/hello-server-volume.yml
deployment.apps/hello-server created
configmap/hello-server-configmap created
```

```sh
❯ kub get pod
NAME                            READY   STATUS    RESTARTS   AGE
hello-server-594ccc7f64-8jsz7   1/1     Running   0          13s
hello-server-594ccc7f64-qx8jx   1/1     Running   0          13s
hello-server-594ccc7f64-tbxjt   1/1     Running   0          13s
```

volumeとしてマウントされていることを確認できます。

```sh
❯ kub describe configmaps hello-server-configmap
Name:         hello-server-configmap
Namespace:    default
Labels:       <none>
Annotations:  <none>

Data
====
myconfig.txt:
----
I am hungry.


BinaryData
====

Events:  <none>
```

port-forwardしてみましょう。

```sh
❯ kub port-forward deployments/hello-server 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

```sh
❯ curl localhost:8080
I am hungry.%
```

#### ConfigMapを壊してみる

まずは、k8s/configmap/hello-server-env.yml をapplyして、DeploymentとConfigMapを作成してみましょう。

```sh
❯ kub apply -f k8s/configmap/hello-server-env.yml
deployment.apps/hello-server created
configmap/hello-server-configmap created
```

次に、k8s/configmap/hello-server-env-destruction.yml をapplyして、ConfigMapを壊してみましょう。

```sh
❯ kub apply -f k8s/configmap/hello-server-destruction.yml
deployment.apps/hello-server configured
configmap/hello-server-configmap unchanged
```

port-forwardしてみましょう。

```sh
❯ kub port-forward deployments/hello-server 8081:8081
error: unable to forward port because pod is not running. Current status=Pending
```

Podを確認してみましょう

```sh
❯ kub get pod
NAME                           READY   STATUS                       RESTARTS   AGE
hello-server-67588987f-bhpz8   0/1     CreateContainerConfigError   0          54s
```
PodのStatusがCreateContainerConfigErrorになっていることがわかります。
このエラーは、コンテナの設定を作成する際に問題が発生したことを示しています。
ConfigMapの内容が正しくないため、Podが正常に起動できないことが原因です。

```sh
❯ kub describe pod
Name:             hello-server-67588987f-bhpz8
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.31.0.2
Start Time:       Sat, 30 May 2026 01:49:51 +0900
Labels:           app=hello-server
                  pod-template-hash=67588987f
Annotations:      <none>
Status:           Pending
IP:               10.244.0.10
IPs:
  IP:           10.244.0.10
Controlled By:  ReplicaSet/hello-server-67588987f
Containers:
  hello-server:
    Container ID:
    Image:          blux2/hello-server:1.4
    Image ID:
    Port:           <none>
    Host Port:      <none>
    State:          Waiting
      Reason:       CreateContainerConfigError
    Ready:          False
    Restart Count:  0
    Environment:
      PORT:  <set to the key 'PORT' of config map 'hello-server-configmap'>  Optional: false
      HOST:  <set to the key 'HOST' of config map 'hello-server-configmap'>  Optional: false
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-mq9s2 (ro)
Conditions:
  Type                        Status
  PodReadyToStartContainers   True
  Initialized                 True
  Ready                       False
  ContainersReady             False
  PodScheduled                True
Volumes:
  kube-api-access-mq9s2:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type     Reason     Age                 From               Message
  ----     ------     ----                ----               -------
  Normal   Scheduled  109s                default-scheduler  Successfully assigned default/hello-server-67588987f-bhpz8 to kind-control-plane
  Normal   Pulled     5s (x10 over 108s)  kubelet            Container image "blux2/hello-server:1.4" already present on machine
  Warning  Failed     5s (x10 over 108s)  kubelet            Error: couldn't find key HOST in ConfigMap default/hello-server-configmap
```

Warningのイベントに、Error: couldn't find key HOST in ConfigMap default/hello-server-configmapとあることがわかります。
次に、deploymentでしているenvのkeyと、ConfigMapのkeyを確認してみましょう。

```sh
❯ kub get deployment hello-server -o yaml
apiVersion: apps/v1
kind: Deployment
.....

      containers:
      - env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              key: PORT
              name: hello-server-configmap
        - name: HOST
          valueFrom:
            configMapKeyRef:
              key: HOST
              name: hello-server-configmap
        image: blux2/hello-server:1.4
        imagePullPolicy: IfNotPresent
        name: hello-server
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
.....
```

```sh
❯ kub get configmaps hello-server-configmap -o yaml
apiVersion: v1
data:
  PORT: "8081"
kind: ConfigMap
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","data":{"PORT":"8081"},"kind":"ConfigMap","metadata":{"annotations":{},"name":"hello-server-configmap","namespace":"default"}}
  creationTimestamp: "2026-05-29T16:48:43Z"
  name: hello-server-configmap
  namespace: default
  resourceVersion: "1584"
  uid: e83ff9a9-de6b-4511-a2fb-c36a535479df
```

ConfigMapにHOSTというkeyが存在しないことがわかります。
ConfigMapを編集して、HOSTというkeyを追加してみましょう。

HOST: "localhost"

```sh
❯ kub apply -f k8s/configmap/hello-server-fixed.yml
deployment.apps/hello-server unchanged
configmap/hello-server-configmap configured
```

PodのStatusがRunningになっていることを確認できます。

```sh
❯ kub get pods
NAME                           READY   STATUS    RESTARTS   AGE
hello-server-67588987f-bhpz8   1/1     Running   0          6m51s
```

## Secretを作成してみよう

データベースのパスワードなどの機密情報を管理するためのリソースです。
SecretはBase64エンコードされたデータを保存しますが、暗号化されているわけではないため、アクセス制御が重要です。
Secretを作成してPodから読み込む方法は、ConfigMapと同様に、環境変数として読み込む方法と、ボリュームとしてマウントして読み込む方法があります。

### 環境変数として読み込む

k8s/secret/nginx-secret.yml に nginxのユーザー名とパスワードをSecretに保存して、Podから読み込んでみましょう。

```sh
❯ kub apply -f k8s/secret/nginx-secret.yml
pod/nginx-sample created
secret/nginx-secret created
```

```sh
❯ kub get pod,secrets
NAME               READY   STATUS              RESTARTS   AGE
pod/nginx-sample   0/1     ContainerCreating   0          10s

NAME                  TYPE     DATA   AGE
secret/nginx-secret   Opaque   2      10s
```

podにアクセスして、環境変数として読み込まれていることを確認できます。

```sh
❯ kub exec -it nginx-sample -- /bin/sh
# echo $USERNAME
admin
# echo $PASSWORD
admin123
# exit
```

### ボリュームとしてマウントして読み込む
k8s/secret/nginx-secret-volume.yml に nginxのユーザー名とパスワードをSecretに保存して、Podから読み込んでみましょう。

```sh
❯ kub apply -f k8s/secret/nginx-volume.yaml
pod/nginx-sample created
secret/nginx-secret created
```

podにアクセスして、ボリュームとしてマウントされていることを確認できます。
```sh
❯ kub exec -it nginx-sample -- /bin/sh
# cat /etc/config/server.key
eM9ku3ecCpUL9zPoIIuG2ptZZC5Cu4ZCQXRymlHajYvZyffpM6
# exit
```
