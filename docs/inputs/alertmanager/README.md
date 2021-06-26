# AlertManager Inputs

AlertManagerを利用する場合の設定について記載します。  

## 前提条件

- AlertManagerからAlertManager Inputsへのネットワーク疎通が可能なこと
- InputsからAutoScaler Coreへの疎通が可能なこと

## AlertManagerの設定

AlertManager Inputsを利用するにはAlertManagerでWebhookの設定が必要です。  

### AlertManagerのWebhook設定

AlertManagerのConfigで以下のように設定します。

```yaml
route:
  receiver: default #任意のデフォルトレシーバー
  routes:
    - receiver: 'autoscaler-up'
      matchers:
        - autoscaler = up # 任意、Prometheus側のアラート設定と揃える

    - receiver: 'autoscaler-down'
      matchers:
        - autoscaler = down # 任意、Prometheus側のアラート設定と揃える

receivers:
  - name: 'default'
    # ... 任意のデフォルトレシーバー

  - name: 'autoscaler-up'
    webhook_configs:
      - url: "AlertManager InputsのURL>/up?[key=value]..." #詳細は下記を参照
        send_resolved: false
        
  - name: "AlertManager InputsのURL>/down?[key=value]..." #詳細は下記を参照
    webhook_configs:
      - url: http://autoscaler-inputs:8080/down
        send_resolved: false

```

- `webhook_configs.url` : `<AlertManager InputsのURL>/up?[key=value]...`

`webhook_configs.url`には以下のパラメータが指定可能です。

- `source`: リクエスト元を識別するための名称。任意の値を利用可能。デフォルト値:`default`
- `action`: 実行するアクション名。Coreのコンフィギュレーションで定義したアクション名を指定する。デフォルト値:`default`
- `resource-group-name`: 操作対象のリソースグループの名前。Coreのコンフィギュレーションで定義したグループ名を指定する。デフォルト値:`default`
- `desired-state-name`: 希望する状態の名前。Coreのコンフィギュレーションで定義したプラン名を指定する。特定の時刻に特定のスペックにしたい場合などに利用する。デフォルト値:`""`  

これらのパラメータを複数指定する場合は`&`で繋げて記載します。  

`webhook_configs.url`の記載例: `http://example.com:8080/up?source=grafana&action=action1&resource-group-name=group1`


## PrometheusのAlert設定

AlertManagerの`routes`で指定したmatchersに合致するようにAlertを設定します。

設定例:

```yaml
groups:
- name: example
  rules:
  - alert: scale_up
    expr: avg_over_time(sakuracloud_server_cpu_time{zone="tk1a"}[1h]) > 80 # tk1aゾーンのサーバのCPU-TIMEの直近一時間の平均
    for: 5m
    labels:
      autoscaler: up # AlertManager側のmatchersと揃える
    annotations:
      summary: "Servers on tk1a zone are keeping high workloads, scaling up..."
      description: "Servers on tk1a zone are keeping high workloads, scaling up..."
```

Note: AutoScalerによる操作でアラート状態が解消できるようなルールを設定してください。  
AutoScaler Coreは同一の`source`/`action`/`resource-group-name`へのリクエストを冷却期間の間は無視しますが、冷却期間がすぎると再度リクエストを受け付けるようになります。  
このためアラートの条件設定次第ではスケール動作を繰り返してしまいます。  

## TLS関連設定

[Inputs共通設定](../tls_config.md)を参照ください。  