# Configuration Reference

sacloud/autoscalerのコンフィギュレーションファイルはYAML形式で書かれた、操作対象のリソース定義や動作の調整を行うためのものです。  

コンフィギュレーションファイルの例:

```yaml
# 操作対象のリソースの定義
resources:
  web: # リソースグループの名前、任意の名称を指定可能
    
    # アクションの定義: Up/Downリクエスト時に指定するアクション名をここで定義する
    # アクションごとにハンドラーを限定したい場合に利用したいハンドラーを指定する
    # actions:
    #   server-vertical-scaling:
    #     - server-vertical-scaler
    #     - elb-servers-handler
    #   elb-vertical-scaling:
    #     - elb-vertical-scaler

    # スケールさせたいリソース群をここで定義する
    resources:
      # サーバ(垂直スケール)
      - type: Server
        selector:
          names: ["example"]
          zone: "is1a"
        option:
          shutdown_force: true

# カスタムハンドラーの定義
# handlers:
#   - type: "fake"
#     name: "fake"
#     endpoint: "unix:autoscaler-handlers-fake.sock"

# オートスケーラーの動作設定
autoscaler:
  job_cooling_sec: 6000 # デフォルト: 6000(10分)
```

## 指定可能な項目

トップレベルには以下の項目を指定します。

- `resources`: 操作したいさくらのクラウド上のリソースのグループ
- `handlers`(省略可): カスタムハンドラーの定義
- `autoscaler`(省略可): オートスケーラー自体の動作設定
- `sakuracloud`(省略可): さくらのクラウドAPIキーなどのさくらのクラウドAPIを利用するための設定

## `resources`

さくらのクラウド上の操作したいリソースのグループを定義します。

形式: ディクショナリ - `グループ名(文字列): リソースグループ(resource_group)`

- グループ名: 任意のグループ名を指定します。ここで指定したグループ名はInputsからのリクエスト時に指定され、操作対象のリソースを選択するのに利用されます。  
- リソースグループ: 複数のリソースの定義 + アクションの組み合わせ。後述の`resource_group`項を参照してください。

### `resource_group`

`resources`の要素として指定します。
以下の項目が指定可能です。  

- `actions`: このリソースグループに対するアクションのリスト。アクション名をキーとし、ハンドラー名のリストを要素として持ちます。
- `resources`: 操作対象のさくらのクラウド上のリソースのリスト。各要素については後述の`resource`を参照してください。

`actions`は省略可能です。省略するとスケーリングに全てのハンドラーが利用されます。  
ここで指定したアクション名はInputsからのリクエスト時に指定され、実行したいハンドラを選択するのに利用されます。

ハンドラー名にはビルトインハンドラーの名前、もしくはトップレベルの`handlers`で指定したカスタムハンドラーの名前を指定可能です。  

- [ResourceGroup](https://pkg.go.dev/github.com/sacloud/autoscaler/core#ResourceGroup)

#### `resource`

`resource_group`の`resources`の要素として指定します。
以下のリソースが指定可能です。

- [DNS](https://pkg.go.dev/github.com/sacloud/autoscaler/core#DNS)
- [EnhancedLoadBalancer](https://pkg.go.dev/github.com/sacloud/autoscaler/core#EnhancedLoadBalancer)
- [GSLB](https://pkg.go.dev/github.com/sacloud/autoscaler/core#GSLB)
- [Router](https://pkg.go.dev/github.com/sacloud/autoscaler/core#Router)
- [Server](https://pkg.go.dev/github.com/sacloud/autoscaler/core#Server)

## `handlers`

カスタムハンドラのリストを指定します。  

- [Handler](https://pkg.go.dev/github.com/sacloud/autoscaler/core#Handler)

## `autoscaler`

オートスケーラー自体の動作設定を行います。

- [AutoScalerConfig](https://pkg.go.dev/github.com/sacloud/autoscaler/core#AutoScalerConfig)

## `sakuracloud`

さくらのクラウドAPIキーなどの設定を行います。

- [SakuraCloud](https://pkg.go.dev/github.com/sacloud/autoscaler/core#SakuraCloud)