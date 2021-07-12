# 汎用Webhook Inputs

汎用Webhook Inputsとは、様々な監視ツールからのWebhookを受け、任意のスクリプトを用いてハンドルすべきWebhookかを判定するInputsです。  

このドキュメントではWebhook Inputsを利用する場合の設定について記載します。  

## 前提条件

- 何らかのWebhookを受信可能であること
- InputsからAutoScaler Coreへの疎通が可能なこと

## 利用例: MackerelのWebhookを利用する場合

[Mackerel](https://mackerel.io)のWebhookでの通知を利用する例です。

まず、Webhookのボディを解析し、AutoScalerでハンドルすべきリクエストなのかを判定するためのスクリプトを作成します。  
参考: [Mackerelヘルプ: Webhookにアラートを通知する](https://mackerel.io/ja/docs/entry/howto/alerts/webhook)  

ここでは`jq`コマンドを利用し、Webhookのボディの`alert.status`が`warning`だった場合はハンドルするようにします。  
以下の内容で`mackerel.sh`ファイルを作成します。

```mackerel.sh
#!/bin/sh

jq -e 'select(.alert.status == "warning")'
```

作成したら`chmod +x`などで適切な権限を付与しておいてください。

```bash
$ chmod +x mackerel.sh
```

次にWebhook Inputsを起動します。

```bash
$ autoscaler inputs webhook --accept-http-methods "POST" --executable-path mackerel.sh
```

## `--executable-path`に指定するスクリプトについて

実行可能なファイルを指定します。スクリプトが終了コード0を返したらハンドルすべきWebhookと判断します。

## TLS関連設定

[Inputs共通設定](../config.md)を参照ください。  