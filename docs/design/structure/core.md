# Core: Structure

AutoScaler Coreが取り扱うデータ構造

### Configuration

AutoScaler起動時に与えられる動作設定など。

```yaml
config: 
  # さくらのクラウド関連コンフィグ
  sakuracloud:
    token: "your-api-key"
    secret: "your-api-secret"
    
  # 公開するアクション: キー:handlersのリストを定義。
  # Inputsからのリクエスト時に任意で指定される
  # いくつかのアクションはデフォルトで提供される
  actions:
    - server_horizontal_scaling:
        - server_horizontal_scaler
        - route53_handler
    - server_vertical_scaling:
        - server_vertical_scaler
    - your_custom_rule:
        - your_custom_handler
    
  # 利用するHandlersの指定  
  handlers:
    # ビルトインHandler: パラメータを受けないものはデフォルトで利用可能
    # - type: server_horizontal_scaler 
    # - type: server_vertical_scaler
    
    # ビルトインHandlers + パラメータあり
    - type: shell_exec_handler
      name: route53_handler
      script_path: /usr/lib/sacloud/auto_scaler/handlers/aws_handler.sh
      work_dir: /usr/lib/sacloud/auto_scaler/handlers/

    # カスタムHandler: gRPCエンドポイントのアドレスを指定する
    - type: external 
      name: your_custom_handler
      endpoint: unix:///var/run/sacloud/your_handler.sock # see https://github.com/grpc/grpc/blob/master/doc/naming.md

  resources: #任意のキーでリソースのグループを定義(Inputsは操作対象としてキーを指定する)
    web: 
      - name: front-servers
        type: ServerGroup
        # ...
      - name: load-balancer
        type: EnhancedLoadBalancer
        # ...
      - name: rdb
        type: Server
        # ...
      
      
# その他動作関連
  
  # TLS関連の設定
  server_tls_config: # Coreがリクエストを受ける際のTLS設定
    # TODO 項目追加
  client_tls_config: # Coreがexternalなhandlersへリクエストする際のTLS設定
    # TODO 項目追加
```

`resources`で指定するリソースについては[Resources](resources.md)を参照。

### Handlersの呼び出し

- Handlersは複数指定可能
- Handlersの呼び出しは順次同期的に行う  
- Handlersの順次呼び出しの都度Handlersに渡すパラメータをリフレッシュする
  -> Handlersからは詳しい処理結果を受け取らない。代わりにCoreがリフレッシュすることで処理結果を次のHandlersに渡す役目をはたす