terraform {
  required_providers {
    sakuracloud = {
      source  = "sacloud/sakuracloud"
      version = "2.8.4"
    }
  }
}

provider "sakuracloud" {
  zone = "is1a"
}

# Server
resource "sakuracloud_server" "server" {
  name   = "autoscaler-e2e-test"
  core   = 1
  memory = 1
  network_interface {
    upstream = "shared"
  }
  force_shutdown = true
}

# ELB
resource "sakuracloud_proxylb" "autoscaler-e2e-test" {
  name    = "autoscaler-e2e-test"
  plan    = 100
  timeout = 10
  region  = "is1"

  health_check {
    protocol   = "tcp"
    delay_loop = 10
    port       = 80
  }

  bind_port {
    proxy_mode = "http"
    port       = 80
    response_header {
      header = "Cache-Control"
      value  = "public, max-age=10"
    }
  }

  server {
    ip_address = sakuracloud_server.server.ip_address
    port       = 80
  }
}