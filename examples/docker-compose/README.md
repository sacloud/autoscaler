# Examples: Docker Composeでの起動例

この例はPrometheus/Grafana/AutoScalerをDocker Composeを用いて起動するものです。  

**注意: この例は開発環境での利用を想定したものです。セキュリティ面の考慮を十分に行っていませんので動作確認用途でのみご利用ください。**

## 準備

以下の環境変数を設定します。

- `SAKURACLOUD_ACCESS_TOKEN`: さくらのクラウドAPIトークン
- `SAKURACLOUD_ACCESS_TOKEN_SECRET`: さくらのクラウドAPIシークレット
- `GF_SECURITY_ADMIN_PASSWORD`: Grafanaの管理者パスワード

   export SAKURACLOUD_ACCESS_TOKEN="your-token"
   export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-secret"
   export GF_SECURITY_ADMIN_PASSWORD="your-password"

次にAutoScalerのコンフィギュレーションを作成/編集します。

    autoscaler core example > autoscaler/autoscaler.yaml
    vi autoscaler/autoscaler.yaml

## 起動

    docker compose up -d
    # grafana
    open http://localhost:3000
    # prometheus
    open http://localhost:9090

### スケールアップ/ダウン

デフォルトでは`watch`ディレクトリ内に以下のファイルが存在するかを検知してスケールアップ/ダウンを行います。  

- `up`: スケールアップ/スケールアウト
- `down`: スケールダウン/スケールイン

`touch watch/up`などで空ファイルを作成してください。