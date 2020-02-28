package main

import (
  "demo/route"
)

const uploadServiceHost  = "127.0.0.1:8080"

func main() {
  router := route.Router()
  router.Run(uploadServiceHost)
}
