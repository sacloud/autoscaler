# Getting Started

sacloud/autoscalerのプロトタイプを動かすための手順

### ソースコードのクローン〜ビルド

```shell
$ git clone https://github.com/sacloud/autoscaler.git; cd autoscaler
$ make
```

### 設定ファイルの作成

```shell
$ cp example-autoscaler.yaml autoscaler.yaml
# 必要に応じて編集
$ vim autoscaler.yaml
```

### autoscalerの起動

```shell
$ bin/autoscaler
```

### (オプション) カスタムハンドラーの起動 

```shell
# Fakeハンドラーを起動する例
$ bin/autoscaler-handlers-fake
```

### スケールアップ/ダウン or スケールアウト/インの実行(手動実行)

コマンドラインから直接スケールアップ/ダウンを実行するためのInputs`autoscaler-inputs-direct`を提供しています。

```shell
# スケールアップ
$ bin/autoscaler-inputs-direct up

# スケールダウン
$ bin/autoscaler-inputs-direct down
```