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
      resources:
        - name: front-servers
          type: ServerGroup
          # ...
        - name: load-balancer
          type: EnhancedLoadBalancer
          # ...
        - name: rdb
          type: Server
          # ...    
      handlers: # (オプション)任意のハンドラーのみ利用したい場合に指定する
        - server_horizontal_scaler  
        - route53_handler
      
      
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