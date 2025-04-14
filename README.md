# AHT21B to InfluxDB

I2C 接続の AHT21B センサーから温度・湿度を取得して、InfluxDB に保存する Go アプリケーションです。Rasberry Pi 4で動作を確認しています。また、このアプリケーションのためのコンテナイメージもGitHub Actionsでbuildして公開しています。

## 構成
- [main.go](main.go): センサー読み取りとデータ送信のロジック
- [Dockerfile](Dockerfile): コンテナイメージのビルド定義
- [example-manifests/aht21b.yaml](example-manifests/aht21b.yaml) / [grafana.yaml](example-manifests/grafana.yaml) / [influxdb.yaml](example-manifests/influxdb.yaml): Kubernetes 用デプロイ例

## 必要な環境変数
- `INFLUXDB_TOKEN` (必須): InfluxDB への書き込みに必要な認証トークン
- `INFLUXDB_URL` (省略可): InfluxDB のURL。指定がない場合は `http://localhost:8086`
- `INFLUXDB_ORG` (省略可): 組織名。指定がない場合は `lemolatoon`
- `INFLUXDB_BUCKET` (省略可): 保存先バケット。指定がない場合は `sensor-home`
- `SLEEP_DURATION_SECONDS` (省略可): センサー読み取りの間隔（秒単位）。デフォルトは `60`

## コンテナイメージ
コンテナは、[ここ](https://github.com/lemolatoon/AHT21BToInfluxDB/pkgs/container/aht21btoinfluxdb)で公開されています。
`ghcr.io/lemolatoon/aht21btoinfluxdb:master`が最新です。

## Kubernetes へのデプロイ
Kubernetes 上で動かす場合は、`example-manifests` フォルダに実際に自分が運用している際のyamlがおいてあります。

PVCには、[local-path-provioner](https://github.com/rancher/local-path-provisioner)を、Ingressには、[cloudflare-tunnel-ingress-controller](https://github.com/STRRL/cloudflare-tunnel-ingress-controller)を使っています。
```sh
kubectl apply -f example-manifests/aht21b.yaml
kubectl apply -f example-manifests/influxdb.yaml
kubectl apply -f example-manifests/grafana.yaml
```