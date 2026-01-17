name = "app"
version = "1.0.0"

config = {
  host = "localhost"
  port = 8080
  ssl  = false
}

servers = [
  {
    name = "server1"
    ip   = "192.168.1.1"
  },
  {
    name = "server2"
    ip   = "192.168.1.2"
  }
]

metadata = {
  tags = ["prod", "web"]
  owner = "team"
}
