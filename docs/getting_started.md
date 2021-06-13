# Getting Started Guide

## 利用までの流れ

- インストール
- Coreの設定ファイルの作成
- Coreの起動
- Inputsの起動
- Grafana/AlertManagerの設定

## インストール

GitHub Releasesから実行ファイルをダウンロードします。  

[sacloud/autoscaler リリースページ](https://github.com/sacloud/autoscaler/releases/latest)

ダウンロードしたら適切なディレクトリに保存/展開してください。

CLIの利用方法については[CLIリファレンス](./cli.md)を参照してください。

## Coreの設定ファイル(autoscaler.yaml)の作成

sacloud/autoscalerを実行するにはYAML形式の設定ファイルで対象リソースの定義などを行う必要があります。  

設定ファイルの雛形は`autoscaler server example`で出力できます。
設定ファイルの記載内容については[Configuration Reference](./configuration.md)を参照してください。

## Coreの起動

以下のコマンドでCoreを起動します。  

```shell
# デフォルト設定で起動
$ autoscaler server start 
```

指定可能なオプションは以下の通りです。

```console
start autoscaler's core server

Usage:
  autoscaler server start [flags]...

Flags:
      --addr string     Address of the gRPC endpoint to listen to (default "unix:autoscaler.sock")
      --config string   File path of configuration of AutoScaler Core (default "autoscaler.yaml")
  -h, --help            help for start

Global Flags:
      --log-format string   Format of logging to be output. options: [ logfmt | json ] (default "logfmt")
      --log-level string    Level of logging to be output. options: [ error | warn | info | debug ] (default "info")
```

Note: Coreはデフォルトだと`unix:autoscaler.sock`でリッスンします。  
Inputsを別のマシン上で動かす場合などは`--addr`フラグで`http://192.0.2.1:8080`のようなアドレスを指定する必要があります。  

## Inputsの起動

以下のコマンドでInputsを起動します。

```shell
# Grafana Inputsの場合
$ autoscaler inputs grafana

# AlertManager Inputsの場合
$ autoscaler inputs alertmanager
```

## Grafana/AlertManagerの設定

Grafana、またはAlertManagerでアラートの設定、およびWebhookでの通知設定が必要です。  
詳細は[Inputsドキュメント](./inputs)を参照してください。  
