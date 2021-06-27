## Exporterについて

AutoScalerではPrometheus向けに各コンポーネントのメトリクスをエクスポートしています。

## Exporterの起動方法

各コンポーネントの設定ファイルにてExporterの設定を記載する必要があります。

### Inputs

Inputsの起動時に`--config`で渡す設定ファイルにて設定します。  

```yaml
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

詳細は[Inputsの共通設定](inputs/config.md)を参照してください。

### Core

Coreのコンフィギュレーションの中で設定します。

```yaml
autoscaler:
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

詳細は[コンフィギュレーション リファレンス](configuration.md)を参照してください。

## メトリクス

以下のメトリクスをサポートしています。

#### gRPC関連

- `sacloud_autoscaler_grpc_errors_total`: gRPCコールのエラー数

ラベルとして呼び出し元 or 呼び出し先のコンポーネント名を持つ

#### Webhook関連

- `sacloud_autoscaler_webhook_requests_total`:  `/up`+`/down`へのリクエスト総数
- `sacloud_autoscaler_webhook_requests_up`: `/up`へのリクエスト数
- `sacloud_autoscaler_webhook_requests_down`: `/down`へのリクエスト数

ラベルとしてステータスコード(200 or 400 or 500)を持つ

### Inputsでの出力例

```
# HELP sacloud_autoscaler_grpc_errors_total The total number of errors
# TYPE sacloud_autoscaler_grpc_errors_total counter
sacloud_autoscaler_grpc_errors_total{component="inputs_to_core"} 0

# HELP sacloud_autoscaler_webhook_requests_down A counter for requests to the /down webhook
# TYPE sacloud_autoscaler_webhook_requests_down counter
sacloud_autoscaler_webhook_requests_down{code="200"} 0
sacloud_autoscaler_webhook_requests_down{code="400"} 0
sacloud_autoscaler_webhook_requests_down{code="500"} 0

# HELP sacloud_autoscaler_webhook_requests_total A counter for requests to the webhooks
# TYPE sacloud_autoscaler_webhook_requests_total counter
sacloud_autoscaler_webhook_requests_total{code="200"} 0
sacloud_autoscaler_webhook_requests_total{code="400"} 0
sacloud_autoscaler_webhook_requests_total{code="500"} 0

# HELP sacloud_autoscaler_webhook_requests_up A counter for requests to the /up webhook
# TYPE sacloud_autoscaler_webhook_requests_up counter
sacloud_autoscaler_webhook_requests_up{code="200"} 0
sacloud_autoscaler_webhook_requests_up{code="400"} 0
sacloud_autoscaler_webhook_requests_up{code="500"} 0
```

### Coreでの出力例

```
# HELP sacloud_autoscaler_grpc_errors_total The total number of errors
# TYPE sacloud_autoscaler_grpc_errors_total counter
sacloud_autoscaler_grpc_errors_total{component="core"} 0
sacloud_autoscaler_grpc_errors_total{component="core_to_handlers"} 0
```