# 汎用Webhook Inputs

汎用Webhook Inputsとは、様々な監視ツールからのWebhookを受け、任意のスクリプトを用いてハンドルすべきWebhookかを判定するInputsです。  

このドキュメントではWebhook Inputsを利用する場合の設定について記載します。  

## 前提条件

- 何らかのWebhookを受信可能であること
- InputsからAutoScaler Coreへの疎通が可能なこと

### Webhook送信側の設定

監視ツールなどでAutoScaler CoreへのUpまたはDownリクエストを送信するためのWebhookを設定します。  

送信先URLは以下のようにします。

- スケールアップ/スケールアウト用のエンドポイント: `<Webhook InputsのURL>/up?[key=value]...`
- スケールダウン/スケールイン用のエンドポイント: `<Webhook InputsのURL>/down?[key=value]...`

URLには以下のパラメータが指定可能です。

- `source`: リクエスト元を識別するための名称。任意の値を利用可能。デフォルト値:`default`
- `resource-name`: 操作対象のリソースの名前。Coreのコンフィギュレーションで定義したリソース名を指定する。デフォルト値:`default`
- `desired-state-name`: 希望する状態の名前。Coreのコンフィギュレーションで定義したプラン名を指定する。特定の時刻に特定のスペックにしたい場合などに利用する。デフォルト値:`""`

これらのパラメータを複数指定する場合は`&`で繋げて記載します。

Urlの記載例: `http://example.com:8080/up?source=grafana&resource-name=resource1`

## 利用例: MackerelのWebhookを利用する場合

[Mackerel](https://mackerel.io)のWebhookでの通知を利用する例です。

あらかじめMackerel側でWebhookを送信するための設定をしておきます。  
参考: [Mackerelヘルプ: Webhookにアラートを通知する](https://mackerel.io/ja/docs/entry/howto/alerts/webhook)  

次に、Webhookのボディを解析し、AutoScalerでハンドルすべきリクエストなのかを判定するためのスクリプトを作成します。  

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