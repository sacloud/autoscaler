# Systemd Unit

このユニットファイルは`/etc/systemd/system`配下に配置して利用します。

## インストール

### 実行ユーザー

実行ユーザーとして`autoscaler`ユーザーが必要です。
以下のように追加します。  

    useradd -s /sbin/nologin -M autoscaler

### `autoscaler`コマンドのインストール先

デフォルトでは`/usr/local/sbin`配下に`autoscaler`をインストールしておく必要があります。  
インストール先を変更している場合は`*.service`に記載されているパスを変更してください。  

### ユニットファイル

`*.service`を`/etc/systemd/system`配下にコピーしてください。

### sysconfigファイル

実行には`/etc/autoscaler/`配下に設定ファイルが必要です。
サンプルファイル`*.config`をコピーして配置し、編集してご利用ください。  

### AutoScalerのコンフィギュレーション

デフォルトでは`etc/autoscaler/autoscaler.yaml`を利用します。  
`/usr/local/sbin/autoscaler core example`コマンドで雛形を出力し作成してください。  

### デフォルトの動作

デフォルトでは以下のように動作します。

- Core: `/var/run/autoscaler/autoscaler.sock`でgRPCサーバをリッスン開始
- Inputs: `:8080`でhttpサーバをリッスン開始

あらかじめ`/var/run/autoscaler`を作成し、`autoscaler`グループに読み書き権限を付与しておいてください。

    sudo mkdir -p /var/run/autoscaler
    sudo chgrp autoscaler /var/run/autoscaler/
    sudo chmod 770 /var/run/autoscaler/

## 実行

### 有効化

    sudo systemctl enable autoscaler_core.service
    # 以下は必要に応じて
    sudo systemctl enable autoscaler_inputs_grafana.service
    sudo systemctl enable autoscaler_inputs_alertmanager.service
    sudo systemctl enable autoscaler_inputs_zabbix.service

### 開始

    sudo systemctl start autoscaler_core.service
    # 以下は必要に応じて
    sudo systemctl start autoscaler_inputs_grafana.service
    sudo systemctl start autoscaler_inputs_alertmanager.service
    sudo systemctl start autoscaler_inputs_zabbix.service

### ログの確認

    sudo journalctl -f -u autoscaler_*
