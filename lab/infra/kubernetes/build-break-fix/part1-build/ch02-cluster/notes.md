# Kubernetes クラスタを構築する

## kind を 使用してクラスターの構築

P57 に 記載されているコマンドは以下の通りです。

```sh
# デフォルトのk8sのイメージの場合
$ kind create cluster

# カスタムイメージを使用する場合
$ kind create cluster --image kindest/node:v1.24.0
```

実際に実行したコマンド
```sh
$ kind create cluster
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.33.1) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Not sure what to do next? 😅  Check out https://kind.sigs.k8s.io/docs/user/quick-start/
```

クラスターの確認
```sh
$ kubectl cluster-info --context kind-kind
```
> ※ 僕のaliasで kubectl を kub として登録しているため、以下のように実行しています。
```sh
$ kub cluster-info --context kind-kind
Kubernetes control plane is running at https://127.0.0.1:63578
CoreDNS is running at https://127.0.0.1:63578/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```
### kubectl の config
staging や production など、複数のクラスターを管理する場合、kubectl の config を使用してクラスターの情報を管理します。
config は ~/.kube/config に保存されいます。
~/.kube/config を確認すると、kind-cluster の情報が追加されていることがわかります。

実際のコードは見せられないのですが、以下のような内容が追加されているはずです。

```yaml
apiVersion: v1
clusters:
- context:
    cluster: kind-kind
    user: kind-kind
  name: kind-kind
current-context: kind-kind
kind: Config
preferences: {}
users:
- name: kind-kind
  user:
    client-certificate-data: <base64 encoded certificate></base64>
    client-key-data: <base64 encoded key>
```
## クラスターの削除

クラスターの削除は以下のコマンドで行います。
```sh
❯ kind delete cluster
Deleting cluster "kind" ...
Deleted nodes: ["kind-control-plane"]
```
