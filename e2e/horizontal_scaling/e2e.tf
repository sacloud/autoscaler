terraform {
  required_providers {
    sakuracloud = {
      source  = "sacloud/sakuracloud"
      version = "2.29.1"
    }
  }
}

provider "sakuracloud" {
  zone = "tk1b"
}

resource "sakuracloud_proxylb" "autoscaler-e2e-test" {
  name    = "autoscaler-e2e-horizontal-scaling"
  plan    = 100
  timeout = 10
  region  = "is1"

  health_check {
    protocol   = "http"
    delay_loop = 10
    path       = "/"
  }

  bind_port {
    proxy_mode = "http"
    port       = 80
    response_header {
      header = "Cache-Control"
      value  = "public, max-age=10"
    }
  }
}
