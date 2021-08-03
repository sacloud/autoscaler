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