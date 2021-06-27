# Grafana Inputs

Grafanaを利用する場合の設定について記載します。  

## 前提条件

- GrafanaからGrafana Inputsへのネットワーク疎通が可能なこと
- Grafana InputsからAutoScaler Coreへの疎通が可能なこと

## Grafanaの設定

Grafana Inputsを利用するにはGrafana側でNotification Channelを設定した上でAlertsを設定する必要があります。  

### Notification Channelの設定

AutoScaler CoreへのUpまたはDownリクエストを送信するためのWebhookをNotification Channelに登録します。  

![notification_channel.png](./images/notification_channel.png)

パラメータは以下のように入力します。  

- Name: 任意
- Type: webhook
- Url: 
  - スケールアップ/スケールアウト用のエンドポイント: `<Grafana InputsのURL>/up?[key=value]...`
  - スケールダウン/スケールイン用のエンドポイント: `<Grafana InputsのURL>/down?[key=value]...`
- Http Method: `POST` | `PUT` (どちらでも可)

Urlには以下のパラメータが指定可能です。

- `source`: リクエスト元を識別するための名称。任意の値を利用可能。デフォルト値:`default`
- `action`: 実行するアクション名。Coreのコンフィギュレーションで定義したアクション名を指定する。デフォルト値:`default`
- `resource-group-name`: 操作対象のリソースグループの名前。Coreのコンフィギュレーションで定義したグループ名を指定する。デフォルト値:`default`
- `desired-state-name`: 希望する状態の名前。Coreのコンフィギュレーションで定義したプラン名を指定する。特定の時刻に特定のスペックにしたい場合などに利用する。デフォルト値:`""`  

これらのパラメータを複数指定する場合は`&`で繋げて記載します。  

Urlの記載例: `http://example.com:8080/up?source=grafana&action=action1&resource-group-name=group1`

#### Config Fileによるプロビジョニングを行う場合の例

```yaml
# /etc/grafana/provisioning/notifiers/example.yaml
notifiers:
  - name: "AutoScaler:Up"
    type: webhook
    org_id: 1
    uid: 1
    is_default: false
    send_reminder: false
    frequency: 30m
    disable_resolve_message: true
    settings:
      autoResolve: true
      httpMethod: "POST"
      uploadImage: false
      url: "http://example.com:8080/up?source=grafana&action=action1&resource-group-name=group1"
```

#### APIで登録する場合の例

GrafanaのAPI(`POST /api/alert-notifications`)を利用してNotification Channelを登録する例:

```console
POST /api/alert-notifications HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk

{
  "name": "AutoScaler: Up",
  "type": "webhook",
  "isDefault": false,
  "sendReminder": false,
  "disableResolveMessage": true,
  "frequency": "",
  "settings": {
    "autoResolve": true,
    "httpMethod": "POST",
    "uploadImage": false,
    "url": "http://example.com:8080/up?source=grafana&action=action1&resource-group-name=group1"
  }
}
```

## GrafanaでのAlert設定

Note: AutoScalerによる操作でアラート状態が解消できるようなルールを設定してください。  
AutoScaler Coreは同一の`source`/`action`/`resource-group-name`へのリクエストを冷却期間の間は無視しますが、冷却期間がすぎると再度リクエストを受け付けるようになります。  
このためアラートの条件設定次第ではスケール動作を繰り返してしまいます。  

## TLS関連設定

[Inputs共通設定](../config.md)を参照ください。  