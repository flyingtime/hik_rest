HIKVISON RESTFUL API
=================
该项目基于go gin的框架，将海康摄像头SDK封装成restlful的形式
https://github.com/gin-gonic/gin

### project运行的需求

* Linux
* golang
* Cgo

### 编译和运行

* git clone https://github.com/xujieasd/hik_rest
* make
* echo "[your project path]/build" >> /etc/ld.so.conf
* echo "[your project path]/build/HCNetSDKCom/" >> /etc/ld.so.conf
* ldconfig
* cd build
* ./make

### restful接口

1.抓拍接口

POST /cameraCapture HTTP/1.1
```
    输入参数
    {
      "address"  : address,
      "port"     : port,
      "userName" : userName,
      "passWord" : passWord，
      "quality"  : quality,
      "size"     : size
    }

    输出
    {
      “registerCode” : code,
      "cephLocation" : location
    }
```
```
    例
    curl -i -X POST "http://xxx:8080/cameraCapture" -H "Content-Type:application/json" -d '{"address":"xxx","port":8000,"userName":"admin","passWord":"abc12345","quality":"0","size":"0"}'
	HTTP/1.1 200 OK
	Content-Type: application/json; charset=utf-8
	Date: Wed, 16 Aug 2017 10:29:51 GMT
	Content-Length: 66
	Connection: keep-alive
	Keep-Alive: timeout=15

    {"captureCode":0,"cephLocation":"xxx.jpg"}
```

2.摄像机联通性

GET /ConnectionTest HTTP/1.1
```
    输入参数
    {
      "address"  : address,
      "port"     : port,
      "userName" : userName,
      "passWord" : passWord，
    }

    输出
    {
      “registerCode” : code,
      "lUserID"      : id
    }
```
```
    例
    curl -i -X GET "http://xxx:8080/ConnectionTest" -H "Content-Type:application/json" -d '{"address":"xxx","port":8000,"userName":"admin","passWord":"abc12345"}'
	HTTP/1.1 200 OK
	Content-Type: application/json; charset=utf-8
	Date: Mon, 11 Sep 2017 10:00:01 GMT
	Content-Length: 30
	Connection: keep-alive
	Keep-Alive: timeout=15

    {"captureCode":0,"lUserID":0}
```

### 参考
* gin: https://github.com/gin-gonic/gin
* hik sdk: http://www.hikvision.com/cn/download_61.html
* Ceph go sdk: http://github.com/aws/aws-sdk-go
* hik sdk cgo: https://github.com/superhacker777/hikka

