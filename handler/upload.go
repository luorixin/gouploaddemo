package handler

import (
  cmn "demo/common"
  cfg "demo/config"
  dblayer "demo/db"
  "demo/meta"
  "demo/mq"
  "demo/store/ceph"
  "demo/store/oss"
  "demo/util"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "strconv"
  "time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))

	} else if r.Method == "POST" {
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data err:%s", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/fileSave/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err:%s", err.Error())
			return
		}
		defer newFile.Close()
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err :%s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

    // 游标重新回到文件头部
    newFile.Seek(0, 0)

    if cfg.CurrentStoreType == cmn.StoreCeph {
      // 文件写入Ceph存储
      data, _ := ioutil.ReadAll(newFile)
      cephPath := "/ceph/" + fileMeta.FileSha1
      _ = ceph.PutObject("userfile", cephPath, data)
      fileMeta.Location = cephPath
    } else if cfg.CurrentStoreType == cmn.StoreOSS {
      // 文件写入OSS存储
      ossPath := "oss/" + fileMeta.FileSha1
      // 判断写入OSS为同步还是异步
      if !cfg.AsyncTransferEnable {
        err = oss.Bucket().PutObject(ossPath, newFile)
        if err != nil {
          fmt.Println(err.Error())
          w.Write([]byte("Upload failed!"))
          return
        }
        fileMeta.Location = ossPath
      } else {
        // 写入异步转移任务队列
        data := mq.TransferData{
          FileHash:      fileMeta.FileSha1,
          CurLocation:   fileMeta.Location,
          DestLocation:  ossPath,
          DestStoreType: cmn.StoreOSS,
        }
        pubData, _ := json.Marshal(data)
        pubSuc := mq.Publish(
          cfg.TransExchangeName,
          cfg.TransOSSRoutingKey,
          pubData,
        )
        if !pubSuc {
          // TODO: 当前发送转移信息失败，稍后重试
        }
      }
    }



		_ = meta.UpdateFileMetaDB(fileMeta)

		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
		  http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
    }else{
      w.Write([]byte("Upload Failed!"))
    }
	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
    return
  }
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
  username := r.Form.Get("username")
  //fileMetas := meta.GetLastFileMetas(limitCnt)
  fileMetas, err := dblayer.QueryUserFileMetas(username, limitCnt)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  data, err := json.Marshal(fileMetas)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  fsha1 := r.Form.Get("filehash")
  fm:=meta.GetFileMeta(fsha1)
  f, err := os.Open(fm.Location)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  defer f.Close()

  data, err :=ioutil.ReadAll(f)
  if err !=nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "application/octect-stream")
  w.Header().Set("content-disposition", "attachment;filename=\""+fm.FileName+"\"")
  w.Write(data)
}

func FileUpdateHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  opType := r.Form.Get("op")
  filesha1 := r.Form.Get("filehash")
  newFileName := r.Form.Get("fileName")

  if opType != "0" {
    w.WriteHeader(http.StatusForbidden)
    return
  }
  if r.Method != "POST"{
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }

  curFileMeta := meta.GetFileMeta(filesha1)
  curFileMeta.FileName = newFileName
  meta.UpdateFileMeta(curFileMeta)

  data, err := json.Marshal(curFileMeta)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(data)
}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  filesha1 := r.Form.Get("filehash")
  fMeta := meta.GetFileMeta(filesha1)
  os.Remove(fMeta.Location)

  meta.RemoveFileMeta(filesha1)

  w.WriteHeader(http.StatusOK)
  w.Write([]byte("delete ok"))
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request){
  r.ParseForm()

  username := r.Form.Get("username")
  filehash := r.Form.Get("filehash")
  filename := r.Form.Get("filename")
  filesize,_ := strconv.Atoi(r.Form.Get("filesize"))

  fileMeta, err := meta.GetFileMetaDB(filehash)
  if err != nil {
    fmt.Println(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  if fileMeta == nil {
    resp := util.RespMsg{
      Code: -1,
      Msg:  "秒传失败，请访问普通上传接口",
    }
    w.Write(resp.JSONBytes())
    return
  }

  suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
  if suc {
    resp := util.RespMsg{
      Code: 0,
      Msg:  "秒传成功",
    }
    w.Write(resp.JSONBytes())
    return
  }else{
    resp := util.RespMsg{
      Code: -2,
      Msg:  "秒传失败，请稍后重试",
    }
    w.Write(resp.JSONBytes())
    return
  }
}
