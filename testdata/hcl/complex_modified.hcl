name = "app"
version = "1.1.0"

config = {
  host = "example.com"
  port = 8443
  ssl  = true
}

servers = [
  {
    name = "server1"
    ip   = "192.168.1.1"
  },
  {
    name = "server2"
    ip   = "192.168.1.2"
  },
  {
    name = "server3"
    ip   = "192.168.1.3"
  }
]

metadata = {
  tags = ["prod", "web", "https"]
  owner = "team"
  region = "us-east-1"
}
