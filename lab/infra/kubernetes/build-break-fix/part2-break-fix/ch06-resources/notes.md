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

