terraform {
  required_providers {
    sakuracloud = {
      source  = "sacloud/sakuracloud"
      version = "2.14.2"
    }
  }
}

provider "sakuracloud" {
  zone = "is1a"
}

# Server
resource "sakuracloud_server" "server" {
  count = 2

  name   = "autoscaler-e2e-test"
  core   = 1
  memory = 1

  network_interface {
    upstream = "shared"
  }

  disks = [sakuracloud_disk.disk[count.index].id]

  disk_edit_parameter {
    hostname        = "autoscaler-e2e-test"
    disable_pw_auth = true
    note {
      id = sakuracloud_note.startupscript.id
    }
  }
}

resource "sakuracloud_note" "startupscript" {
  name    = "autoscaler-e2e-test"
  content = file("startup-script.sh")
}

resource "sakuracloud_disk" "disk" {
  count             = 2
  name              = "autosxaler-e2e-test"
  source_archive_id = data.sakuracloud_archive.ubuntu.id
}

data "sakuracloud_archive" "ubuntu" {
  os_type = "ubuntu2004"
}

# ELB
resource "sakuracloud_proxylb" "autoscaler-e2e-test" {
  name    = "autoscaler-e2e-test"
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

  dynamic "server" {
    for_each = sakuracloud_server.server
    content {
      ip_address = server["value"].ip_address
      port       = 80
    }
  }
}
