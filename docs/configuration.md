# Configuration Reference

sacloud/autoscalerのコンフィギュレーションファイルはYAML形式で書かれた、操作対象のリソース定義や動作を調整するためのものです。  

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
      # GSLB(配下のサーバが垂直スケールする際にサーバのデタッチ&アタッチが行われる)
      - type: GSLB
        selector:
          names: ["example-gslb"]
        resources:
          # サーバ(垂直スケール)
          - type: Server
            selector:
              names: ["example"]
              zones: ["is1a"]
            option:
              shutdown_force: true

# カスタムハンドラーの定義
# handlers:
#   - name: "fake"
#     endpoint: "unix:autoscaler-handlers-fake.sock"

# オートスケーラーの動作設定
autoscaler:
  cooldown: 600 

# さくらのクラウド API関連
# sakuracloud:
#   token: "<your-token>"
#   secret: "<your-secret>"
```

## スキーマ

```yaml
# オートスケール対象リソースの定義
resources:
  <resource_def_groups>

# カスタムハンドラーのリスト(省略可能)
handlers: 
  [ - <handler> ]

# オートスケーラーの動作設定(省略可能)
autoscaler:
  <autoscaler_config>

# さくらのクラウドAPI関連設定(省略可能)
# Note: APIキーにはアクセスレベル`作成・削除`または`設定編集`が必要です
sakuracloud:
  # APIトークン、省略した場合は環境変数SAKURACLOUD_ACCESS_TOKENを参照します
  token: <string>
  # APIシークレット、省略した場合は環境変数SAKURACLOUD_ACCESS_TOKEN_SECRETを参照します
  secret:  <string>
```

## 各要素の詳細

- [\<resource_def_groups\>](#resource_def_groups)
  - [\<resource_def_group\>](#resource_def_group)
    - [\<action\>](#action)
    - [\<resource_definition\>](#resource_definition) 
      - [\<resource_def_dns\>](#resource_def_dns)
      - [\<resource_def_elb\>](#resource_def_elb)
      - [\<resource_def_gslb\>](#resource_def_gslb)
      - [\<resource_def_load_balancer\>](#resource_def_load_balancer)
      - [\<resource_def_router\>](#resource_def_router)
      - [\<resource_def_server\>](#resource_def_server)
- [\<handler\>](#handler)
- [\<autoscaler_config\>](#autoscaler_config)
- 

### \<resource_def_groups\>

操作したいさくらのクラウド上のリソースのグループ(リソースグループ)を定義します。
グループ名とresource_def_groupのmapとして指定します。

ここで指定したグループ名はInputsから処理対象のリソースを指定するパラメータとして利用されます。  

```yaml
{ <string> : <resource_def_group> }
```

### \<resource_def_group\>

このリソースグループに対するアクションと対象リソースを定義します。

```yaml
# アクションのリスト(省略可能)
actions: 
  [ - <action> ]

# 操作対象リソース定義のリスト(必須)
resources:
  [ - <resource_definition> ]
```

### \<action\>

リソースグループに対し、有効にするハンドラーの組み合わせを定義します。  
アクション名とハンドラ名のリストのmapとして指定します。
ここで指定したアクション名はInputsから処理対象のアクションを指定するパラメータとして利用されます。  
ハンドラー名にはビルトインハンドラーの名前、もしくはトップレベルの\<handlers\>で指定したカスタムハンドラーの名前を指定可能です。

```yaml
{ <string> : [ - <string> ] }
```

### \<resource_definition\>

操作対象となるさくらのクラウド上のリソースの定義
resource_def_xxxのいずれかを指定します。

```yaml
<resource_def_dns> | <resource_def_elb> | <resource_def_gslb> | <resource_def_load_balancer> | <resource_def_router> | <resource_def_server>
```

### \<resource_def_dns\>

DNSリソースの定義

```yaml
type: "DNS"
selector:
  # idかnamesのどちらかを指定、必須
  id: <string | number>
  # 部分一致、複数指定した場合はAND結合
  names: 
    [ - <string> ]

# 子リソースの定義(省略可能)
resources:
  [ - <resource_definition> ]
```

### \<resource_def_elb\>

エンハンスドロードバランサの定義。  
ここで定義したリソースは垂直スケール可能になります(ハンドラ`elb-vertical-scaler`)。  
また、`resources`配下に\<resource_def_server\>を定義し、かつELBの配信先サーバとIPアドレスが一致するサーバについては
水平スケールする前にELBからのデタッチ/アタッチが行われます(ハンドラ:`elb-servers-handler`)。

```yaml
type: "ELB" # or EnhancedLoadBalancer
selector:
  # idかnamesのどちらかを指定、必須
  id: <string | number>
  # 部分一致、複数指定した場合はAND結合
  names:
    [ - <string> ]
  # 垂直スケールさせる範囲(省略可能)
  plans:
    [ - name: <string> # プラン名、省略可能 
        cps: <number> ]
# 子リソースの定義(省略可能)
resources:
  [ - <resource_definition> ]
```

`plans`を省略した場合のデフォルト値は以下の通りです。

```yaml
# ELBのデフォルトの垂直スケール範囲
plans:
  - cps: 100
  - cps: 500
  - cps: 1000
  - cps: 5000
  - cps: 10000
  - cps: 50000
  - cps: 100000
  - cps: 400000
```

### \<resource_def_gslb\>

GSLBの定義。
`resources`配下に\<resource_def_server\>を定義し、かつGSLBの宛先サーバとIPアドレスが一致するサーバについては
水平スケールする前にGSLBからのデタッチ/アタッチが行われます(ハンドラ:`gslb-servers-handler`)。

```yaml
type: "GSLB"
selector:
  # idかnamesのどちらかを指定、必須
  id: <string | number>
  # 部分一致、複数指定した場合はAND結合
  names: 
    [ - <string> ]

# 子リソースの定義(省略可能)
resources:
  [ - <resource_definition> ]
```

### \<resource_def_load_balancer\>

ロードバランサの定義。
`resources`配下に\<resource_def_server\>を定義し、かつLBの宛先サーバとIPアドレスが一致するサーバについては
水平スケールする前にLBからのデタッチ/アタッチが行われます(ハンドラ:`load-balancer-servers-handler`)。

```yaml
type: "LoadBalancer"
selector:
  # idかnamesのどちらかを指定、必須
  id: <string | number>
  # 部分一致、複数指定した場合はAND結合
  names: 
    [ - <string> ]
  zones:
    [ - <"is1a" | "is1b" | "tk1a" | "tk1b" | "tk1v"> ]

# 子リソースの定義(省略可能)
resources:
  [ - <resource_definition> ]
```

### \<resource_def_router\>

ルータの定義。  
ここで定義したリソースは垂直スケール可能になります(ハンドラ`router-vertical-scaler`)。  

```yaml
type: "Router" 
selector:
  # idかnamesのどちらかを指定、必須
  id: <string | number>
  # 部分一致、複数指定した場合はAND結合
  names:
    [ - <string> ]
  zones:
    [ - <"is1a" | "is1b" | "tk1a" | "tk1b" | "tk1v"> ]
  # 垂直スケールさせる範囲(省略可能)
  plans:
    [ - name: <string> # プラン名、省略可能 
        band_width: <number> ]
# 子リソースの定義(省略可能)
resources:
  [ - <resource_definition> ]
```

`plans`を省略した場合のデフォルト値は以下の通りです。

```yaml
# ルータのデフォルトの垂直スケール範囲
plans:
  - band_width: 100
  - band_width: 250
  - band_width: 500
  - band_width: 1000
  - band_width: 1500
  - band_width: 2000
  - band_width: 2500
  - band_width: 3000
  - band_width: 3500
  - band_width: 4000
  - band_width: 4500
  - band_width: 5000
```

### \<resource_def_server\>

サーバの定義。  
ここで定義したリソースは垂直スケール可能になります(ハンドラ`server-vertical-scaler`)。

```yaml
type: "Server" 
selector:
  # idかnamesのどちらかを指定、必須
  id: <string | number>
  # 部分一致、複数指定した場合はAND結合
  names:
    [ - <string> ]
  zones:
    [ - <"is1a" | "is1b" | "tk1a" | "tk1b" | "tk1v"> ]
  
  # コア専有プランを利用するか
  dedicated_cpu: <boolean>
  
  # 垂直スケール動作のオプションを指定(省略可能)
  option:
    # 強制シャットダウンを行うか(ACPIが利用できないサーバの場合trueにする)
    shutdown_force: <boolean> 
  
  # 垂直スケールさせる範囲(省略可能)
  plans:
    [ - name: <string> # プラン名、省略可能 
        core: <number> # コア数
        memory: <number> #メモリサイズ、GB単位 
    ]
  
# 子リソースの定義(省略可能)
resources:
  [ - <resource_definition> ]
```

`plans`を省略した場合のデフォルト値は以下の通りです。

```yaml
# サーバのデフォルトの垂直スケール範囲
plans:
  - core: 2
    memory: 4
  - core: 4
    memory: 8
  - core: 4
    memory: 16
  - core: 8
    memory: 16
  - core: 10
    memory: 24
  - core: 10
    memory: 32
  - core: 10
    memory: 48
```

### \<handler\>

カスタムハンドラーの定義。

```yaml
name: <string> #ハンドラ名
endpoint: <string> #gRPCのエンドポイントアドレス(例: unix:/var/run/your-custom-handler.sock)
```

### \<autoscaler_config\>

オートスケーラーの動作設定

```yaml
cooldown: <number | default = 600> # 同一ジョブの連続実行を抑制するためのクールダウン期間、秒単位で指定

# CoreのgRPCエンドポイントのTLS関連設定
server_tls_config:
  cert_file: <string> # 証明書のファイルパス
  key_file: <string> # 秘密鍵のファイルパス
  # クライアント認証タイプ: 詳細は https://golang.org/pkg/crypto/tls/#ClientAuthType を参照
  client_auth_type: <"NoClientCert" | "RequestClientCert" | "RequireAnyClientCert" | "VerifyClientCertIfGiven" | "RequireAndVerifyClientCert" >
  client_ca_file: <string> # クライアント認証で利用するCA証明書(チェイン)のファイルパス

# HandlersへのgRPCリクエスト時のTLS関連設定
handler_tls_config:
  cert_file: <string> # 証明書のファイルパス
  key_file: <string> # 秘密鍵のファイルパス 
  root_ca_file: <string> # ルート証明書(チェイン)のファイルパス

# Exporterの設定
exporter_config:
  enabled: <bool>
  address: <string | default=":8081">
  tls_config:
    cert_file: <string> # 証明書のファイルパス
    key_file: <string> # 秘密鍵のファイルパス
    # クライアント認証タイプ: 詳細は https://golang.org/pkg/crypto/tls/#ClientAuthType を参照
    client_auth_type: <"NoClientCert" | "RequestClientCert" | "RequireAnyClientCert" | "VerifyClientCertIfGiven" | "RequireAndVerifyClientCert" >
    client_ca_file: <string> # クライアント認証で利用するCA証明書(チェイン)のファイルパス
```

## 名前付きプラン

ELB/ルータ/サーバでは名前付きプランを利用可能です。  

```yaml
# サーバのでの名前付きプランの例
plans:
  - core: 2
    memory: 4
    name: "low"
  - core: 4
    memory: 8
  - core: 4
    memory: 16
    name: "middle"
  - core: 8
    memory: 16
  - core: 10
    memory: 24
    name: "high"
  - core: 10
    memory: 32
  - core: 10
    memory: 48
```

プラン名はInputsからの`DesiredStateName`パラメータで指定されます。  
Coreは`DesiredStateName`が指定されていると以下のようにあるべき姿を算出します。  

- Upリクエストがきた場合、現在のリソースのプランより大きな名前付きプランを返す
  見つからなかった場合はプラン変更せずエラーとする
  
- Downリクエストがきた場合、現在のリソースのプランより小さな名前付きプランを返す
  見つからなかった場合はプラン変更せずエラーとする
