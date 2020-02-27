package main

import (
  "bufio"
  "demo/config"
  "demo/mq"
  "encoding/json"
  "log"
  "os"
)

func ProcessTransfer(msg []byte) bool {
  // 解析msg
  pubData := mq.TransferData{}
  err := json.Unmarshal(msg, pubData)
  if err !=nil {
    log.Println(err.Error())
    return false
  }

  // 根据临时存储路径创建文件句柄
  filed, err := os.Open(pubData.CurLocation)
  if err != nil {
    log.Println(err.Error())
    return false
  }

  reader := bufio.NewReader(filed)

  println(reader)
  // 通过文件句柄读取上传到oss
  return true
}

func main() {
  if !config.AsyncTransferEnable {
    log.Println("异步转移文件功能目前被禁用，请检查相关配置")
    return
  }
  log.Println("文件转移服务启动中，开始监听转移任务队列...")
  mq.StartConsume(
    config.TransOSSQueueName,
    "transfer_oss",
    ProcessTransfer)
}
