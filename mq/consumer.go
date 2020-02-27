package mq

import "log"

var done chan bool

func StartConsume(qName, cName string, callback func(msg []byte) bool) {
  // 获取消息信道
  msgs, err := channel.Consume(qName, cName, true, false, false, false, nil)
  if err != nil {
    log.Println(err.Error())
    return
  }
  // 循环获取队列消息
  done = make(chan bool)
  go func() {
    for msg := range msgs {
      // 调用callback处理消息
      processSuc := callback(msg.Body)
      if !processSuc {
        //TODO 写到另一个队列用于异常情况的重试
      }
    }
  }()
  // done 没有新消息过来，会一直阻塞
  <-done

  channel.Close()
}
