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

