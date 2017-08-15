package main

/*
#include "HCNetSDK.h"
#include <stdlib.h>
*/
import "C"
import (
	"github.com/gin-gonic/gin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service"
	//"github.com/golang/glog"
	//"github.com/spf13/pflag"
	"net/http"
	"unsafe"
	//"flag"
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)


type Login struct {
	User     string `form:"user" json:"user" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type CameraLogin struct {
	Address     string `form:"address" json:"address" binding:"required"`
	Port        int    `form:"port" json:"port" binding:"required"`
	UserName    string `form:"userName" json:"userName" binding:"required"`
	PassWord    string `form:"passWord" json:"passWord" binding:"required"`
}

type CameraCapture struct {
	Size     int `form:"size" json:"size" binding:"required"`
	Quality  int `form:"quality" json:"quality" binding:"required"`

}

type CephMgr struct {
	Host		string
	Bucket		string
	AccessKey 	string
	SecretKey	string
	PathStyle       bool
}

var Ceph = &CephMgr{
	AccessKey: "",
	SecretKey: "",
	Host:      "",
	Bucket:    "",
	PathStyle: true,
}



func (CM *CephMgr) connect()(*session.Session, error){

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:           aws.String("default"),
			Endpoint:         &CM.Host,
			S3ForcePathStyle: &CM.PathStyle,
			Credentials:      credentials.NewStaticCredentials(CM.AccessKey,CM.SecretKey,""),
		},
	}))

	return sess, nil
}

func (CM *CephMgr) upload(fileName string, key string) error{

	sess, _ := Ceph.connect();

	file, err:= os.Open(fileName)
	if err != nil{
		fmt.Printf("error: upload: cannot open file %v\n", err)
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(CM.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil{

		fmt.Printf("Unable to upload %q to %q\n", fileName, CM.Bucket,)
		fmt.Println(err.Error())

		return err
	}
	fmt.Printf("log: upload: success upload file %q to %q, %v\n", fileName, CM.Bucket, err)

	return nil

}

func (CM *CephMgr) listBucket(){
	sess, _ := Ceph.connect();
	svc:= s3.New(sess)

	result, err := svc.ListBuckets(nil)

	if err != nil{
		fmt.Printf("list fail %v\n",err)
	}else{
		for _,b := range result.Buckets{
			fmt.Printf("%s create on %s\n",aws.StringValue(b.Name),aws.TimeValue(b.CreationDate))
		}
	}
}

func HikLogin(ip string, port int, login string, password string) (int,int64) {
	var device C.NET_DVR_DEVICEINFO

	c_ip := C.CString(ip)
	defer C.free(unsafe.Pointer(c_ip))

	c_login := C.CString(login)
	defer C.free(unsafe.Pointer(c_login))

	c_password := C.CString(password)
	defer C.free(unsafe.Pointer(c_password))

	uid := (int64)(C.NET_DVR_Login(
		c_ip,
		C.WORD(port),
		c_login,
		c_password,
		(*C.NET_DVR_DEVICEINFO)(unsafe.Pointer(&device)),
	))

	if uid >= 0 {

		fmt.Printf("log: Logged in: %s:%s@%s id is %d\n", login, password, ip, uid)

		//C.NET_DVR_Logout((C.LONG)(uid))

		return 0, uid
	} else {
		fmt.Printf("error: Logged error: %s:%s@%s id is %d\n", login, password, ip, uid)
		return (int)(C.NET_DVR_GetLastError()), uid
	}
}

func checkLogin(ip string, port int, login string, password string) bool {
	var device C.NET_DVR_DEVICEINFO

	c_ip := C.CString(ip)
	defer C.free(unsafe.Pointer(c_ip))

	c_login := C.CString(login)
	defer C.free(unsafe.Pointer(c_login))

	c_password := C.CString(password)
	defer C.free(unsafe.Pointer(c_password))

	uid := (int64)(C.NET_DVR_Login(
		c_ip,
		C.WORD(port),
		c_login,
		c_password,
		(*C.NET_DVR_DEVICEINFO)(unsafe.Pointer(&device)),
	))

	if uid >= 0 {

		fmt.Printf("log: Logged in: %s:%s@%s id is %d\n", login, password, ip, uid)
		processSnapshots(ip, uid, login, password, device)

		C.NET_DVR_Logout((C.LONG)(uid))

		return true
	} else {
		return false
	}
}

func getIpChannelsCount(uid int64) (count int) {
	var ipcfg C.NET_DVR_IPPARACFG
	var written int32

	// Getting count of IP cams
	if C.NET_DVR_GetDVRConfig(
		(C.LONG)(uid),
		C.NET_DVR_GET_IPPARACFG,
		0,
		(C.LPVOID)(unsafe.Pointer(&ipcfg)),
		(C.DWORD)(unsafe.Sizeof(ipcfg)),
		(*C.uint32_t)(unsafe.Pointer(&written)),
	) >= 0 {
		for i := 0; i < C.MAX_IP_CHANNEL && ipcfg.struIPChanInfo[i].byEnable == 1; i++ {
			count++
		}
	}

	return
}

func processSnapshots(ip string, uid int64, login string, password string, device C.NET_DVR_DEVICEINFO) {
	ip_count := getIpChannelsCount(uid)

	// SHIT
	if ip_count != 0 || device.byChanNum != 0 {
		if device.byChanNum != 0 {
			getSnapshots(
				ip,
				uid,
				(int)(device.byStartChan),
				(int)(device.byChanNum),
				login,
				password,
			)
		}

		if ip_count != 0 {
			getSnapshots(
				ip,
				uid,
				(int)(device.byStartChan)+32,
				(int)(ip_count),
				login,
				password,
			)
		}
	} else {
		fmt.Printf("warn: No cameras on %s\n", ip)
	}
}

func getSnapshots(ip string, uid int64, startChannel int, count int, login string, password string) {
	downloaded := 0
	//var shoots_path = "~/temp"
	for i := startChannel; i < startChannel+count; i++ {
		//filename := fmt.Sprintf("%s%s_%s_%s_%d.jpg", shoots_path, login, password, ip, i)
		filename := fmt.Sprintf("%s_%s_%s_%d.jpg", login, password, ip, i)
		c_filename := C.CString(filename)
		defer C.free(unsafe.Pointer(c_filename))

		var imgParams C.NET_DVR_JPEGPARA
		imgParams.wPicQuality = 0
		imgParams.wPicSize = 0

		result := C.NET_DVR_CaptureJPEGPicture(
			(C.LONG)(uid),
			(C.LONG)(i),
			(*C.NET_DVR_JPEGPARA)(unsafe.Pointer(&imgParams)),
			c_filename,
		)

		if result == 0 {

			fmt.Printf("error: Error while downloading a snapshot from %s:%d\n", ip, (int)(C.NET_DVR_GetLastError()))
		} else {
			os.Chmod(filename, 0644)
			downloaded++
		}
	}

	if downloaded != 0 {
		fmt.Printf("log: Downloaded %d photos from %s\n", downloaded, ip)
	} else {
		fmt.Printf("warn: Can't get photos from %s\n", ip)
	}
}

func main() {
	// Disable Console Color
	// gin.DisableConsoleColor()
	C.NET_DVR_Init()
	defer C.NET_DVR_Cleanup()
	//pflag.Parse()
	//flag.Set("logtostderr", "true")
	var nLogLevel int
	var strLogDir string
	nLogLevel = 3
	strLogDir = "./sdkLog"

	Ceph.listBucket()
	//err := Ceph.upload("test.jpg","test-8-15.jpg")

	C.NET_DVR_SetLogToFile(C.DWORD(nLogLevel), C.CString(strLogDir),C.int(1));
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})


	r.POST("/loginJSON", func(c *gin.Context) {
		var json Login

		if c.BindJSON(&json) == nil {
			if json.User == "manu" && json.Password == "123" {
				checkLogin("10.19.138.110",8000,"admin","abc12345")

				c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			}
		}
	})

	r.GET("/cameraLogin", func(c *gin.Context) {
		var cameraLogin CameraLogin

		if c.BindJSON(&cameraLogin) == nil {
			var result, uid = HikLogin(cameraLogin.Address,cameraLogin.Port,cameraLogin.UserName,cameraLogin.PassWord)
			if result == 0 {

				c.JSON(http.StatusOK, gin.H{"registerCode": 0, "lUserID": uid})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"registerCode": result, "lUserID": 0})
			}
		}
	})

	r.GET("/cephUpload", func(c *gin.Context) {
		err := Ceph.upload("test.jpg","test-8-15.jpg")
		if err == nil{
			c.JSON(http.StatusOK, gin.H{"status": "done"})
		}else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed"})
		}

	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	/*
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			DB[user] = json.Value
			c.JSON(200, gin.H{"status": "ok"})
		}
	})
	*/
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}