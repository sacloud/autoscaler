# Examples: Terraformとの相互運用

この例はTerraformで作成/管理するリソースをオートスケールの対象とするものです。
`local-exec`カスタムハンドラーを用いてPostHandleのタイミングでTerraformのステートファイルを更新(リフレッシュ)します。

## 使い方

例としてELBの垂直スケールを行います。

### Terraformでのリソースの作成

以下のようなtfファイルでELBを作成します。
ELBのプランは垂直スケールで変更されるため、`lifecycle`を指定してプランをTerraform管理外とします。  

```tf
terraform {
  required_providers {
    sakuracloud = {
      source  = "sacloud/sakuracloud"
      version = "2.11.0"
    }
  }
}

resource sakuracloud_proxylb "elb" {
  name           = "with-terraform"
  plan           = 100
  vip_failover   = true
  sticky_session = true
  gzip           = true

  health_check {
    protocol = "http"
    path     = "/healthz"
  }

  bind_port {
    proxy_mode = "http"
    port       = 80
  }

  # planは垂直スケールで変更されるためTerraformでの管理外とする
  lifecycle {
    ignore_changes = [
      plan
    ]
  }
}
```

### `local-exec`カスタムハンドラーの準備

以下のようなシェルスクリプトを用意します。
PostHandle時は`result`にCREATED/UPDATED/DELETED/NOOPのいずれかが渡されるため、jqコマンドで`result`の値を用いて処理すべきか判定します。  

```bash
#!/bin/bash

# 標準入力経由で渡されるJSONの例
#{
#  "source": "example",
#  "resource_name": example"
#  "scaling_job_id": "example",
#  "result": 1, # CREATED:1 | UPDATED:2 | DELETED:3
#  "current": {
#    "Resource": ...
#  }
#}

# CREATED/UPDATED/DELETEDの場合はterraformのステートをリフレッシュする
SHOULD_HANDLE=$(jq '[.] | any(.result == 1 or .result == 2 or .result == 3)')

if [ "$SHOULD_HANDLE" == "true" ]; then
  WORKING_DIR=$(cd $(dirname $0); pwd)
  terraform -chdir="$WORKING_DIR" apply -refresh-only -auto-approve 2>&1 1> /dev/null
  exit $?
else
  echo "ignored"
  exit 0
fi
```

以下のようにして起動しておきます。

```bash
$ autoscaler handlers local-exec --executable-path refresh.sh --handler-type post-handle
```

### Coreの起動

以下のように`local-exec`をカスタムハンドラーとして登録しておきます。

```yaml
# リソースの定義
resources:
  - type: ELB
    name: "elb"
    selector:
      names: ["with-terraform"]
      
# カスタムハンドラーの定義
handlers:
  - name: "local-exec"
    endpoint: "unix:autoscaler-handlers-local-exec.sock"
```

後はスケールアップ/ダウンを行うとTerraformのステートファイルが更新されます。  
