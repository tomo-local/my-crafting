# アプリケーションのヘルスチェック

- 3種類のProbes
  - Liveness Probe: コンテナが生きているかどうかを確認するためのプローブ。コンテナが応答しない場合、Kubernetesはコンテナを再起動します。
  - Readiness Probe: コンテナがリクエストを受け付ける準備ができているかどうかを確認するためのプローブ。コンテナが準備できていない場合、Kubernetesはそのコンテナへのトラフィックを停止します。
  - Startup Probe: コンテナの起動時に使用されるプローブ。コンテナが起動してから一定時間内に応答しない場合、Kubernetesはコンテナを再起動します。

## Readiness Probe

Podがリクエストを受け付ける準備ができているかどうかを確認するためのプローブ。Podが準備できていない場合、KubernetesはそのPodへのトラフィックを停止します。

pod-readiness.yaml を apply して、確認してみましょう。
imageでは、最初の5秒間はリクエストに応答しないように設定しています。
その後に、Podが準備できているかどうかを確認するためのReadiness Probeが定義されています。

```sh
❯ kub apply -f pod-readiness.yaml
```

watch コマンドでPodの状態を確認してみましょう。
最初は問題はなく、PodはRunning状態になりますが、Readiness Probeが失敗するため、Podは準備できていない状態になります。

```sh
❯ kub get pod --watch
NAME                    READY   STATUS    RESTARTS   AGE
http-server-readiness   1/1     Running   0          11s
http-server-readiness   0/1     Running   0          26s
```

Podが準備できていない状態になると、KubernetesはそのPodへのトラフィックを停止します。
リクエストのlogを確認してみましょう。

```sh
❯ kub logs http-server-readiness
2026/05/31 16:16:41 Starting server...
2026/05/31 16:16:46 Health Check: OK
2026/05/31 16:16:51 Health Check: OK
2026/05/31 16:16:56 Error: Service Unhealthy
2026/05/31 16:17:01 Error: Service Unhealthy
2026/05/31 16:17:06 Error: Service Unhealthy
2026/05/31 16:17:06 Error: Service Unhealthy
```

Podが準備できていない状態になると、リクエストに対してエラーが返されるようになります。
このように、Readiness Probeを使用することで、Podがリクエストを受け付ける準備ができているかどうかを確認することができます。

## Liveness Probe

Readiness Probeと似ていますが、違う点があります。
それは、最後の失敗時の動作です。Readiness Probeは、Podへのトラフィックを停止しますが、Liveness Probeは、Podを再起動をおこないます。
Podが再起動して、問題が解決されケースの場合はこちらの方が有効です。ですが、Podが再起度しても失敗してしまう場合は、再起動の繰り返しになってしまう可能性があります。
なので、安易にLiveness Probeを使用することは、おすすめしません。Readiness Probeを使用して、Podが準備できていない状態になる原因を特定することが重要です。
なので、構成としては最初にReadiness Probeを使用して、Podが準備できていない状態になる原因を特定し、その後にLiveness Probeを使用して、Podが再起動するようにすることが望ましいです。
yamlの構成としては、Readiness ProbeとLiveness Probeを両方定義して、それぞれの実行タイミングを調整して運用することが望ましいです。

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5

livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

pod-liveness.yaml を apply して、確認してみましょう。
imageでは、最初の5秒間はリクエストに応答しないように設定しています。
その後に、Podが生きているかどうかを確認するためのLiveness Probeが定義されています。

```sh
❯ kub apply -f pod-liveness.yaml
```

```sh
❯ kub get pod --watch
NAME                   READY   STATUS    RESTARTS      AGE
http-server-liveness   1/1     Running   4 (24s ago)   2m4s
http-server-liveness   0/1     CrashLoopBackOff   4 (1s ago)    2m6s
```

probeがFailedして、Podが再起動されていることがわかります。
describeコマンドで、Podの状態を確認してみましょう。
```sh
❯ kub describe pod http-server-liveness
Name:             http-server-liveness
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.31.0.2
Start Time:       Mon, 01 Jun 2026 01:24:27 +0900
Labels:           app=http-server
Annotations:      <none>
Status:           Running
IP:               10.244.0.17
IPs:
  IP:  10.244.0.17
Containers:
  http-server:
    Container ID:   containerd://b3b1bad34c51af052cc2a666986bd9684b64c4245562424e842adf502e2da1b2
    Image:          blux2/delayfailserver:1.1
    Image ID:       docker.io/blux2/delayfailserver@sha256:84c46dd90117eda4f2545504e8ce9b2e595eef9fedb02aa2e0dcaa0c13cfeba0
    Port:           <none>
    Host Port:      <none>
    State:          Waiting
      Reason:       CrashLoopBackOff
    Last State:     Terminated
      Reason:       Error
      Exit Code:    2
      Started:      Mon, 01 Jun 2026 01:27:21 +0900
      Finished:     Mon, 01 Jun 2026 01:27:42 +0900
    Ready:          False
    Restart Count:  5
    Liveness:       http-get http://:8080/healthz delay=5s timeout=1s period=5s #success=1 #failure=3
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-6hxbr (ro)
Conditions:
  Type                        Status
  PodReadyToStartContainers   True
  Initialized                 True
  Ready                       False
  ContainersReady             False
  PodScheduled                True
Volumes:
  kube-api-access-6hxbr:
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
  Type     Reason     Age                     From               Message
  ----     ------     ----                    ----               -------
  Normal   Scheduled  4m11s                   default-scheduler  Successfully assigned default/http-server-liveness to kind-control-plane
  Normal   Pulled     2m56s (x4 over 4m11s)   kubelet            Container image "blux2/delayfailserver:1.1" already present on machine
  Normal   Created    2m56s (x4 over 4m11s)   kubelet            Created container http-server
  Normal   Started    2m56s (x4 over 4m11s)   kubelet            Started container http-server
  Normal   Killing    2m56s (x3 over 3m46s)   kubelet            Container http-server failed liveness probe, will be restarted
  Warning  Unhealthy  2m41s (x10 over 3m56s)  kubelet            Liveness probe failed: HTTP probe failed with statuscode: 503
```

## Startup Probe
Startup Probeは、コンテナの起動時に使用されるプローブです。コンテナが起動してから一定時間内に応答しない場合、Kubernetesはコンテナを再起動します。
Startup Probeは、コンテナの起動に時間がかかる場合や、起動時に一時的な問題が発生する可能性がある場合に有効です。Startup Probeを使用することで、コンテナが正常に起動するまでの時間を確保することができます。
Startup Probeは、v1.18から追加された機能で、今まではReadiness ProbeやLiveness Probeを使用して、コンテナの起動時の状態を管理していました。

下記のマニュフェストでは、最大30秒間*10回のリトライを行うように設定しています。これにより、コンテナが起動するまでの時間を確保することができます。

```yaml
startupProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

# 実際にハンズオンで確認してみましょう。

deployment-destruction.yamlをapplyして、Deploymentを作成してみましょう。

```sh
❯ kub apply -f deployment-destruction.yaml
```

watch コマンドでPodの状態を確認してみましょう。
```sh
❯ kub get pod --watch
NAME                            READY   STATUS              RESTARTS   AGE
hello-server-5bc7ccb8dd-kxdl2   0/2     ContainerCreating   0          3s
hello-server-5bc7ccb8dd-mm8sz   0/2     ContainerCreating   0          3s
hello-server-5bc7ccb8dd-zccqc   0/2     ContainerCreating   0          3s
hello-server-5bc7ccb8dd-zccqc   1/2     Running             0          14s
hello-server-5bc7ccb8dd-mm8sz   1/2     Running             0          15s
hello-server-5bc7ccb8dd-kxdl2   1/2     Running             0          16s
```

PodがStatusがRunningになっていますが、READYは1/2となっています。
では、Podの状態を確認してみましょう。
```sh
❯ kub describe pod hello-server-5bc7ccb8dd-kxdl2
Name:             hello-server-5bc7ccb8dd-kxdl2
Namespace:        default
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.31.0.2
Start Time:       Mon, 01 Jun 2026 01:36:20 +0900
Labels:           app=hello-server
                  pod-template-hash=5bc7ccb8dd
Annotations:      <none>
Status:           Running
IP:               10.244.0.20
IPs:
  IP:           10.244.0.20
Controlled By:  ReplicaSet/hello-server-5bc7ccb8dd
Containers:
  hello-server:
    Container ID:   containerd://ce4f0478c82c4340d87fe2cc45ae19a485b67898fa3d32fd6c63b8fb27e6dadc
    Image:          blux2/hello-server:1.6
    Image ID:       docker.io/blux2/hello-server@sha256:035c114efa5478a148e5aedd4e2209bcc46a6d9eff3ef24e9dba9fa147a6568d
    Port:           8080/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Mon, 01 Jun 2026 01:36:28 +0900
    Ready:          False
    Restart Count:  0
    Liveness:       http-get http://:8080/health delay=10s timeout=1s period=5s #success=1 #failure=3
    Readiness:      http-get http://:8081/health delay=5s timeout=1s period=5s #success=1 #failure=3
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-ngrzx (ro)
  busybox:
    Container ID:  containerd://bedb1fa774d42c9b5cb079f72a0dd6b32586d8d9fbc389951a9975db9ac4a712
    Image:         busybox:1.36.1
    Image ID:      docker.io/library/busybox@sha256:73aaf090f3d85aa34ee199857f03fa3a95c8ede2ffd4cc2cdb5b94e566b11662
    Port:          <none>
    Host Port:     <none>
    Command:
      sleep
      9999
    State:          Running
      Started:      Mon, 01 Jun 2026 01:36:35 +0900
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-ngrzx (ro)
Conditions:
  Type                        Status
  PodReadyToStartContainers   True
  Initialized                 True
  Ready                       False
  ContainersReady             False
  PodScheduled                True
Volumes:
  kube-api-access-ngrzx:
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
  Type     Reason     Age                  From               Message
  ----     ------     ----                 ----               -------
  Normal   Scheduled  2m7s                 default-scheduler  Successfully assigned default/hello-server-5bc7ccb8dd-kxdl2 to kind-control-plane
  Normal   Pulling    2m6s                 kubelet            Pulling image "blux2/hello-server:1.6"
  Normal   Pulled     119s                 kubelet            Successfully pulled image "blux2/hello-server:1.6" in 1.509s (7.46s including waiting)
  Normal   Created    119s                 kubelet            Created container hello-server
  Normal   Started    119s                 kubelet            Started container hello-server
  Normal   Pulling    119s                 kubelet            Pulling image "busybox:1.36.1"
  Normal   Pulled     112s                 kubelet            Successfully pulled image "busybox:1.36.1" in 1.26s (7.137s including waiting)
  Normal   Created    112s                 kubelet            Created container busybox
  Normal   Started    112s                 kubelet            Started container busybox
  Warning  Unhealthy  42s (x17 over 110s)  kubelet            Readiness probe failed: Get "http://10.244.0.20:8081/health": dial tcp 10.244.0.20:8081: connect: connection refused
```

Readiness ProbeがFailedしていることがわかります。
このPodは、hello-serverというコンテナとbusyboxというコンテナの2つのコンテナで構成されています。

では、次にlogを確認してみましょう。
```sh
❯ kub logs hello-server-5bc7ccb8dd-kxdl2 -c hello-server
2026/05/31 16:36:28 Starting server on port 8080
2026/05/31 16:36:40 Health Status OK
2026/05/31 16:36:45 Health Status OK
2026/05/31 16:36:50 Health Status OK
2026/05/31 16:36:55 Health Status OK
2026/05/31 16:37:00 Health Status OK
2026/05/31 16:37:05 Health Status OK
2026/05/31 16:37:10 Health Status OK
2026/05/31 16:37:15 Health Status OK
2026/05/31 16:37:20 Health Status OK
2026/05/31 16:37:25 Health Status OK
2026/05/31 16:37:30 Health Status OK
2026/05/31 16:37:35 Health Status OK
2026/05/31 16:37:40 Health Status OK
2026/05/31 16:37:45 Health Status OK
2026/05/31 16:37:50 Health Status OK
2026/05/31 16:37:55 Health Status OK
2026/05/31 16:38:00 Health Status OK
2026/05/31 16:38:05 Health Status OK
2026/05/31 16:38:10 Health Status OK
2026/05/31 16:38:15 Health Status OK
2026/05/31 16:38:20 Health Status OK
2026/05/31 16:38:25 Health Status OK
```

hello-serverコンテナは、8080ポートで起動していますが、Readiness Probeは8081ポートで確認しているため、接続が拒否されていることがわかります。
このように、Readiness ProbeがFailedしている場合は、Podがリクエストを受け付ける準備ができていない状態になっていることがわかります。
修正しましょう

```sh
❯ kub edit deployments hello-server
deployment.apps/hello-server edited
```

修正後問題が解決されているか確認してみましょう。
```sh
❯ kub get pod --watch
NAME                            READY   STATUS        RESTARTS   AGE
hello-server-5bc7ccb8dd-kxdl2   1/2     Terminating   0          7m5s
hello-server-5bc7ccb8dd-mm8sz   1/2     Terminating   0          7m5s
hello-server-5bc7ccb8dd-zccqc   1/2     Terminating   0          7m5s
hello-server-64bf989855-2h5dx   2/2     Running       0          19s
hello-server-64bf989855-8mr6n   2/2     Running       0          30s
hello-server-64bf989855-z4xfx   2/2     Running       0          14s
```

Podが再起動されて、READYが2/2になっていることがわかります。
