## Inputs共通設定

各Inputsは`--config`というパラメータでTLS関連やExporterなどの設定ファイルへのパスを指定可能です。

利用例:

```bash
$ autoscaler inputs grafana --addr 192.0.2.1:443 --config your-config.yaml
```

設定ファイルの書式は以下の通りです。

```yaml
# InputsのWebサーバのエンドポイント向けTLS設定
server_tls_config:
  cert_file: <string> # 証明書のファイルパス
  key_file: <string> # 秘密鍵のファイルパス
  # クライアント認証タイプ: 詳細は https://golang.org/pkg/crypto/tls/#ClientAuthType を参照
  client_auth_type: <"NoClientCert" | "RequestClientCert" | "RequireAnyClientCert" | "VerifyClientCertIfGiven" | "RequireAndVerifyClientCert" >
  client_ca_file: <string> # クライアント認証で利用するCA証明書(チェイン)のファイルパス

# CoreへのgRPCリクエスト時のTLS関連設定
core_tls_config:
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