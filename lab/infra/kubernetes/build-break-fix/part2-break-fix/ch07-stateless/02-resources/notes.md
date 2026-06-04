# アプリケーションに適切なリソースを指定しよう

確保したいリソースの最低使用量を指定することができます。Kubernetesのスケジューラはこの値を見て、スケジュールするNodeを決定します。リソースの要求は、Podがスケジュールされるために必要なリソースの量を指定します。
Requestsは、Podがスケジュールされるために必要なリソースの量を指定します。RequestsのリソースがどのNodeにもない場合、Podはスケジュールされません。
リソースの制限は、Podが使用できるリソースの最大量を指定します。
Limitsは、Podが使用できるリソースの最大量を指定します。LimitsのリソースがPodのRequestsよりも小さい場合、Podはスケジュールされません。


```yaml
resources:
  requests:
    cpu: "500m"
    memory: "128Mi"
  limits:
    cpu: "1"
    memory: "256Mi"
```

## リソースの単位

### メモリ
単位を指定しない場合、デフォルトは1は1bytesになります。
他にも、K, M, G, T, P, Eなどの単位を使用できます。これらはそれぞれ、10の累乗に基づいています。
K = 10の3乗
M = 10の6乗
ほかにも、Ki, Mi, Gi, Ti, Pi, Eiなどの単位も使用できます。これらはそれぞれ、1024の累乗に基づいています。
Ki = 2の10乗
Mi = 2の20乗

### CPU
CPUの単位はコア数を基準にしています。例えば、1は1コアを意味します。
500mは0.5コアを意味します。つまり、500mは500ミリCPUを意味します。
0.5は0.5コアを意味します。つまり、0.5は500mと同じです。

## Quality of Service (QoS) クラス
リソース設定に関して、QoSは覚えておいた方がいいです。
Nodeのメモリが完全に枯渇してしまうと、そのNodeに載っている全てのコンテナが動作できなくなります。Kubernetesは、Nodeのリソースが枯渇したときに、どのPodを優先的に削除するかを決定するために、QoSクラスを使用します。それをOOM Killerという役割の機能が担っています。
QoSに基づいて、OOM KillerするPodの優先順位を決定して、必要に応じて優先度の低いPodをOOM Killします。

QoSクラスには、3種類あります。
- Guaranteed:
  - Podのすべてのコンテナがリソースのrequestsとlimitsを同じ値に設定している場合
- Burstable:
  - Podの少なくとも1つのコンテナがリソースのrequestsとlimitsを異なる値に設定している場合
- BestEffort:
  - Podのすべてのコンテナがリソースのrequestsを設定していない場合

優先順位は、Guaranteed > Burstable > BestEffortの順になります。つまり、Nodeのリソースが枯渇したときに、BestEffortクラスのPodが最初に削除され、その後BurstableクラスのPodが削除され、最後にGuaranteedクラスのPodが削除されます。

## 壊すまたしてもPodが壊れた

最初の `./deployment-resource-handson.yaml` を apply
このマニュフェストではPodは起動しない

1.は リソースの定義でNodeのリソースを超える場合を想定したもの
対応は、まず Pod/Deployment のリソースの定義を見直して、Nodeのリソースを超えないようにすることです。例えば、requestsとlimitsを適切な値に設定することで、PodがNodeのリソースを超えないようにできます。

2.は OMM Kill が発生した場合を想定したもの
今回の対応はimageを修正して、リソースを使用しないものに変更して対応しました
