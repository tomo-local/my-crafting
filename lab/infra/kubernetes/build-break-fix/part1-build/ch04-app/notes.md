# 最小構成リソース: Pod

参考ファイル：./nginx.yml
Podは複数のコンテナーをまとめて起動することができるリソースですが、今回は1つのコンテナーだけを起動する構成にします。
メインのサービス（コンテイナー）に付属するサービス（コンテイナー）のことをsidecar containerと呼びます。
今回はnginxを起動するだけのPodを作成します。
PodはKubernetesの最小構成リソースで、単一のコンテナーを起動することができます。

## リソースを作成するための場所：Namespace
リソースの名前は、Namespace内で一意である必要があります。
Namespace間では、同じ名前のリソースを作成することができます。

# 実際のリソースの作成
まず、kindでclusterを作成します。
一旦今回はローカル環境で動かすことを想定しているため、kindを使用します。
```sh
kind create cluster
```
今回は ./myapp.yaml を使用して、Podを作成します。

applyを実行する前に、Podが存在しないことを確認しておきましょう。

```sh
❯ kub get pods
No resources found in default namespace.
```
applyを実行して、Podを作成します。
```sh
❯ kub apply -f ./myapp.yml -n default
pod/myapp created
```
実際にPodが作成されたことを確認します。
```sh
❯ kub get pods
NAME    READY   STATUS    RESTARTS   AGE
myapp   1/1     Running   0          30s
```
Podが作成され、正常に起動していることがわかります。


# コラム
なぜ kubectl run ではなく、kubectl apply を使用するのか？

run でも、コンテナーを起動することができます。
- マニュフェストがあった方が、リソースの構成を明確にすることができる
- kub runは Podの「冗長化などの高度な設定には使えないため,applyを使用することが推奨されている

kub run は一時的なデバックなどに使われることが多いいらしい
