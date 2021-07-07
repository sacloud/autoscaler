## Inputs: AlertManager Inputs

### 公開するWebhook向けエンドポイント

Grafana Inputsと同様

### Webhookのハンドリング

AlertManagerが送信するWebhookの構造は以下の通り。

参照: [Prometheusのドキュメント:webhook](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)

```json
{
  "version": "4",
  "groupKey": <string>,              // key identifying the group of alerts (e.g. to deduplicate)
  "truncatedAlerts": <int>,          // how many alerts have been truncated due to "max_alerts"
  "status": "<resolved|firing>",
  "receiver": <string>,
  "groupLabels": <object>,
  "commonLabels": <object>,
  "commonAnnotations": <object>,
  "externalURL": <string>,           // backlink to the Alertmanager.
  "alerts": [
    {
      "status": "<resolved|firing>",
      "labels": <object>,
      "annotations": <object>,
      "startsAt": "<rfc3339>",
      "endsAt": "<rfc3339>",
      "generatorURL": <string>       // identifies the entity that caused the alert
    },
    ...
  ]
}
```

- `status`が`firing`の場合にCoreを呼び出すようにし、それ以外は無視する。   
- Coreの`Up`または`Down`を呼び出す。
    - パラメータは以下のように指定
        - `source` = `{{ .Inputsのインスタンス名 }}-{{ .groupKey }}` # TODO 要確認
        - `action` = クエリストリングを参照、未指定の場合は`default`
        - `resource-name` = クエリストリングを参照、未指定の場合は`default`

- `status`が`firing`なリクエストを受け取る都度Coreを呼び出す
    - Core側で同じalert ruleからのwebhookかを`source`で判定し重複処理を防ぐ
    
Note: Coreが処理を行ったらアラートが解消されるようにルールを定めるのは利用者側の責務。  