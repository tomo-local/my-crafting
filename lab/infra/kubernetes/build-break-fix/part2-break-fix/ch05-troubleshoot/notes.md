# トラブルシューティングガイド

## Podの情報を取得する

まずは、Podを取得する
```sh
❯ kub get pod  myapp -n default
NAME    READY   STATUS    RESTARTS   AGE
myapp   1/1     Running   0          78m```

-o wideオプションを付けると、どのノードで動いているかもわかる

```sh
❯ kubectl get pods -o wide -n default
NAME    READY   STATUS    RESTARTS   AGE   IP           NODE                 NOMINATED NODE   READINESS GATES
myapp   1/1     Running   0          79m   10.244.0.5   kind-control-plane   <none>           <none>
```

-o yamlオプションを付けると、Podの詳細な情報がわかる

```sh
❯ kubectl get pod myapp -o yaml -n default
apiVersion: v1
kind: Pod
metadata:
....
```

ファイルに実際のPodの情報を出力することもできる

```sh
❯ kub get pod myapp -o yaml -n default > pod.yml
```

diffで実際に applyしたファイルと、実際のPodの情報を比較することもできる

```sh
❯ diff pod.yml myapp-pod.yml
```
なぜ差分が多く出るのか、実際にymlに記載した内容だけでは、KubernetesはPodを作成できないからである。
Kubernetesは、Podを作成するために、必要な情報を自動で補完しているため、実際のPodの情報と、ymlファイルの内容には差分が出る


-o で jsonpathオプションを付けると、特定の情報だけを抜き取ることもできる

```sh
❯ kub get pod myapp -o jsonpath='{.metadata.name}' -n default
myapp
```
他にもjqのコマンドを使っても特定の情報だけを抜き取ることができる

```sh
❯ kub get pod myapp -o json -n default | jq '.metadata.name'
"myapp"
```

## リソースの詳細を取得する： kubectl describe

getよりも、詳細な情報が必要な場合に、describeコマンドを使いましょう

```sh
❯ kub describe pod myapp
Name:             myapp
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.31.0.2
Start Time:       Mon, 25 May 2026 00:31:39 +0900
Labels:           app=myapp
Annotations:      <none>
Status:           Running
IP:               10.244.0.5
IPs:
  IP:  10.244.0.5
Containers:
  hello-server:
    Container ID:   containerd://2ff962094ac4ea9eb6439700965eef8ff94fa675dd51fec004836bf1b06a1075
    Image:          blux2/hello-server:1.0
    Image ID:       docker.io/blux2/hello-server@sha256:35ab584cbe96a15ad1fb6212824b3220935d6ac9d25b3703ba259973fac5697d
    Port:           8080/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Mon, 25 May 2026 00:31:45 +0900
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-5s2td (ro)
Conditions:
  Type                        Status
  PodReadyToStartContainers   True
  Initialized                 True
  Ready                       True
  ContainersReady             True
  PodScheduled                True
Volumes:
  kube-api-access-5s2td:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:                      <none>
```


## コンテナのログを確認： kubectl logs
使い方に関して、 get で Pod名を取得してから、logsコマンドでログを確認することができます

```sh
❯ kub get pod myapp -n default
NAME    READY   STATUS    RESTARTS   AGE
myapp   1/1     Running   0          78m
```

```sh
❯ kub logs myapp -n default
2026/05/24 15:31:45 Starting server on port 8080
```

特定のDeploymentに紐づくPodのログを確認することもできます

```sh
❯ kub logs deploy/myapp -n default
2026/05/24 15:31:45 Starting server on port 8080
```


## デバック用のサイドカーコンテナを立ち上げる: kubectl debug
今までの参照系のコマンドでは、情報が足りない場合は、kubectl debugコマンドを使って、デバック用のサイドカーコンテナを立ち上げることができます

> kubectl debug --stdin --tty <Pod名> --image=<デバック用イメージ> --target=<コンテナ名> -n default -- sh

```sh
❯ kub debug  --stdin --tty myapp --image=curlimages/curl:8.4.0 --target=hello-server -n default -- sh
```

これで、myappのhello-serverコンテナに対して、curlimages/curl:8.4.0のイメージを使って、デバック用のサイドカーコンテナが立ち上がります。
```sh
If you don't see a command prompt, try pressing enter.
/ # curl localhost:8080
Hello, World!
```

## コンテナを即座に実行する: kubectl run


まぁ、ほかにもいろいろなコマンドがあることがわかった。。。


# 今回のトラブルシューティングの内容は、Podが起動しないというものです。

何回か取得しました。
```sh
❯ kub get pod -n default
NAME    READY   STATUS         RESTARTS   AGE
myapp   0/1     ErrImagePull   0          105m
```

```sh
❯ kub get pod -n default
NAME    READY   STATUS             RESTARTS   AGE
myapp   0/1     CrashLoopBackOff   0          105m
```

この時点でimageのpullに失敗していることがわかりましたが、describeコマンドでさらに詳細な情報を取得してみました

```sh
❯ kub describe pod myapp
Name:             myapp
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.31.0.2
Start Time:       Mon, 25 May 2026 00:31:39 +0900
Labels:           <none>
Annotations:      <none>
Status:           Running
IP:               10.244.0.5
IPs:
  IP:  10.244.0.5
Containers:
...
```

この時点で、イメージのpullに失敗していることがわかりました
なので、実際にimageが存在するのかを確認して、存在しないことがわかりました。
次に、imageを1.1から1.0に修正しました。
修正はeditコマンドを使って、直接Podの定義を修正しました。

```sh
❯ kub edit pod myapp
pod/myapp edited
```
これで、Podの定義を直接修正することができます。
修正後、Podが正常に起動することがわかりました。

```sh
❯ kub get pod -n default
NAME    READY   STATUS    RESTARTS        AGE
myapp   1/1     Running   1 (4m19s ago)   108m
```
