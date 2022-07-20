terraform {
  required_providers {
    sakuracloud = {
      source  = "sacloud/sakuracloud"
      version = "2.17.1"
    }
  }
}

provider "sakuracloud" {
  zone = "is1a"
}

# Server
resource "sakuracloud_server" "autoscaler-e2e-all-resource-types-vertical-server" {
  name   = "autoscaler-e2e-all-resource-types-vertical-server"
  core   = 1
  memory = 2

  network_interface {
    upstream = "shared"
  }

  force_shutdown = true
}

# ELB
resource "sakuracloud_proxylb" "autoscaler-e2e-all-resource-types-elb" {
  name    = "autoscaler-e2e-all-resource-types-elb"
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

resource "sakuracloud_internet" "autoscaler-e2e-all-resource-types-router" {
  name = "autoscaler-e2e-all-resource-types-router"

  netmask     = 28
  band_width  = 100
  enable_ipv6 = false
}