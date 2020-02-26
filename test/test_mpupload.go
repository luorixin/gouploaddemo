package main

import (
  "bufio"
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "net/url"
  "os"
  "strconv"

  jsonit "github.com/json-iterator/go"
)

func multipartUpload(filename string, targetURL string, chunkSize int) error {
  f, err := os.Open(filename)
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer f.Close()

  bfRd := bufio.NewReader(f)
  index := 0

  ch := make(chan int)
  buf := make([]byte, chunkSize) //每次读取chunkSize大小的内容
  for {
    n, err := bfRd.Read(buf)
    if n <= 0 {
      break
    }
    index++

    bufCopied := make([]byte, 5*1048576)
    copy(bufCopied, buf)

    go func(b []byte, curIdx int) {
      fmt.Printf("upload_size: %d\n", len(b))

      resp, err := http.Post(
        targetURL+"&index="+strconv.Itoa(curIdx),
        "multipart/form-data",
        bytes.NewReader(b))
      if err != nil {
        fmt.Println(err)
      }

      body, er := ioutil.ReadAll(resp.Body)
      fmt.Printf("%+v %+v\n", string(body), er)
      resp.Body.Close()

      ch <- curIdx
    }(bufCopied[:n], index)

    //遇到任何错误立即返回，并忽略 EOF 错误信息
    if err != nil {
      if err == io.EOF {
        break
      } else {
        fmt.Println(err.Error())
      }
    }
  }

  for idx := 0; idx < index; idx++ {
    select {
    case res := <-ch:
      fmt.Println(res)
    }
  }

  return nil
}

func main() {
  username := "admin"
  token := "271325c833bec6362411d3b9f4e38b855e5528d7"
  filehash := "4c7012e915a60193865e47c5b7347c61837c0df7"

  // 1. 请求初始化分块上传接口
  resp, err := http.PostForm(
    "http://localhost:8080/file/mpupload/init",
    url.Values{
      "username": {username},
      "token":    {token},
      "filehash": {filehash},
      "filesize": {"41293232"},
    })

  if err != nil {
    fmt.Println(err.Error())
    os.Exit(-1)
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(-1)
  }

  // 2. 得到uploadID以及服务端指定的分块大小chunkSize
  uploadID := jsonit.Get(body, "data").Get("UploadID").ToString()
  chunkSize := jsonit.Get(body, "data").Get("ChunkSize").ToInt()
  fmt.Printf("uploadid: %s  chunksize: %d\n", uploadID, chunkSize)

  // 3. 请求分块上传接口
  filename := "/tmp/fileSave/1.1.5.415.publish_release.dmg"
  tURL := "http://localhost:8080/file/mpupload/uppart?" +
    "username=admin&token=" + token + "&uploadid=" + uploadID
  multipartUpload(filename, tURL, chunkSize)

  // 4. 请求分块完成接口
  resp, err = http.PostForm(
    "http://localhost:8080/file/mpupload/complete",
    url.Values{
      "username": {username},
      "token":    {token},
      "filehash": {filehash},
      "filesize": {"41293232"},
      "filename": {"1.1.5.415.publish_release.dmg"},
      "uploadid": {uploadID},
    })

  if err != nil {
    fmt.Println(err.Error())
    os.Exit(-1)
  }

  defer resp.Body.Close()
  body, err = ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(-1)
  }
  fmt.Printf("complete result: %s\n", string(body))
}
