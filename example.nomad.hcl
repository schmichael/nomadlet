job "j" {
  type        = "batch"
  datacenters = ["dc1"]
  
  constraint {
    attribute = "${attr.nomadlet.version}"
    operator  = "is_set"
  }

  group "g" {
  
    network {
      mode = "none"
    }

    task "t" {
      driver = "raw_exec"

      config {
        command = "echo"
        args    = ["Hello World!"]
      }

      resources {
        cpu    = 100
        memory = 100
      }
    }
  }
}
