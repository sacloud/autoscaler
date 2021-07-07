## Inputs: Grafana Inputs

### 公開するWebhook向けエンドポイント

Grafana Inputsは起動すると以下2つのWebhook向けエンドポイントを公開する。  

- http[s]://example.com/your-inputs-instance-name/up
- http[s]://example.com/your-inputs-instance-name/down

両エンドポイントともパラメータ`resource-name`を指定する必要がある。  

指定例: `https://example.com/my-grafana-input/up?resource-name=web`

- `resource-name`: 処理対象のリソースグループ名、CoreのConfigurationの`resources`で指定したキーを指定する

### Webhookのハンドリング

Grafanaが送信するWebhookの構造は以下の通り。

参照: [Grafanaのドキュメント:webhook](https://grafana.com/docs/grafana/latest/alerting/notifications/#webhook)

```json
{
  "dashboardId":1,
  "evalMatches":[
    {
      "value":1,
      "metric":"Count",
      "tags":{}
    }
  ],
  "imageUrl":"https://grafana.com/assets/img/blog/mixed_styles.png",
  "message":"Notification Message",
  "orgId":1,
  "panelId":2,
  "ruleId":1,
  "ruleName":"Panel Title alert",
  "ruleUrl":"http://localhost:3000/d/hZ7BuVbWz/test-dashboard?fullscreen\u0026edit\u0026tab=alert\u0026panelId=2\u0026orgId=1",
  "state":"alerting",
  "tags":{
    "tag name":"tag value"
  },
  "title":"[Alerting] Panel Title alert"
}
```

- `state`が`alerting`の場合にCoreを呼び出すようにし、それ以外は無視する。   
  参考: stateがとりうる値: `ok`, `paused`, `alerting`, `pending`, `no_data`

- Coreの`Up`または`Down`を呼び出す。
  - パラメータは以下のように指定
    - `source` = `{{ .Inputsのインスタンス名 }}-{{ .dashbordId }}-{{ .orgId }}-{{ .panelId }}-{{ .ruleId }}`
    - `action` = クエリストリングを参照、未指定の場合は`default`
    - `resource-name` = クエリストリングを参照、未指定の場合は`default`

- `state`が`alerting`なリクエストを受け取る都度Coreを呼び出す
  - Core側で同じalert ruleからのwebhookかを`source`で判定し重複処理を防ぐ
  

Note: Coreが処理を行ったらアラートが解消されるようにルールを定めるのは利用者側の責務。  