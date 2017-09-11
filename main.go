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
	"time"
	"strconv"
)


type CameraLogin struct {
	Address     string `form:"address" json:"address" binding:"required"`
	Port        string `form:"port" json:"port" binding:"required"`
	UserName    string `form:"userName" json:"userName" binding:"required"`
	PassWord    string `form:"passWord" json:"passWord" binding:"required"`
}

type CameraLogout struct {
	LUid        string `form:"lUid" json:"lUid" binding:"required"`
}

type CameraCapture struct {
	Address     string `form:"address" json:"address" binding:"required"`
	Port        string `form:"port" json:"port" binding:"required"`
	UserName    string `form:"userName" json:"userName" binding:"required"`
	PassWord    string `form:"passWord" json:"passWord" binding:"required"`
	Quality     string `form:"quality" json:"quality" binding:"required"`
	Size        string `form:"size" json:"size" binding:"required"`

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

func cephErrorf(msg string, args ... interface{})  {
	fmt.Fprintf(os.Stderr,msg+"\n",args...)

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
		ACL:    aws.String("public-read"),
		Body:   file,
	})

	if err != nil{

		//fmt.Printf("Unable to upload %q to %q\n", fileName, CM.Bucket)
		//fmt.Println(err.Error())
		cephErrorf("Unable to upload %q to %q, %v", fileName, CM.Bucket, err)
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

func HikLogout(uid int64){

	C.NET_DVR_Logout((C.LONG)(uid))
	return
}

func ConnectionTest(ip string, port int, login string, password string) (int,int64){
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

		C.NET_DVR_Logout((C.LONG)(uid))

		return 0, uid
	} else {
		fmt.Printf("error: Logged error: %s:%s@%s id is %d\n", login, password, ip, uid)
		return (int)(C.NET_DVR_GetLastError()), uid
	}
}

func loginAndCapture(
	ip string,
	port int,
	login string,
	password string,
	quality int,
	size int,
) (int,string) {
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
		t:= time.Unix(time.Now().Unix(),0)
		filename := fmt.Sprintf("%s_%d_%d_%d_%d_%d_%d.jpg",ip,t.Year(),t.Month(),t.Day(),t.Hour(),t.Minute(),t.Second())
		returnCode := processSnapshots(ip, uid, login, password, quality, size, filename, device)

		C.NET_DVR_Logout((C.LONG)(uid))

		if returnCode == 0 {

			err := Ceph.upload(filename,filename)
			if err !=nil {
				return -1, ""
			}
		}

		return returnCode, filename
	} else {
		return (int)(C.NET_DVR_GetLastError()), ""
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

func processSnapshots(
	ip string,
	uid int64,
	login string,
	password string,
	quality int,
	size int,
	filename string,
	device C.NET_DVR_DEVICEINFO,
) int {
	ip_count := getIpChannelsCount(uid)

	// SHIT
	if ip_count != 0 || device.byChanNum != 0 {
		var result int
		if device.byChanNum != 0 {
			result = getSnapshots(
				ip,
				uid,
				(int)(device.byStartChan),
				(int)(device.byChanNum),
				login,
				password,
				quality,
				size,
				filename,
			)
		}

		if ip_count != 0 {
			result = getSnapshots(
				ip,
				uid,
				(int)(device.byStartChan)+32,
				(int)(ip_count),
				login,
				password,
				quality,
				size,
				filename,
			)
		}
		return result
	} else {
		fmt.Printf("warn: No cameras on %s\n", ip)
		return 3005
	}
}

func getSnapshots(
	ip string,
	uid int64,
	startChannel int,
	count int,
	login string,
	password string,
	quality int,
	size int,
	filename string,
) int {
	downloaded := 0
	//var shoots_path = "~/temp/"
	for i := startChannel; i < startChannel+count; i++ {
		//filename := fmt.Sprintf("%s%s_%s_%s_%d.jpg", shoots_path, login, password, ip, i)
		//filename := fmt.Sprintf("%s_%s_%s_%d.jpg", login, password, ip, i)


		c_filename := C.CString(filename)
		defer C.free(unsafe.Pointer(c_filename))

		var imgParams C.NET_DVR_JPEGPARA
		imgParams.wPicQuality = (C.WORD)(quality)
		imgParams.wPicSize = (C.WORD)(size)

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
	return (int)(C.NET_DVR_GetLastError())
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
	r.GET("/testPing", func(c *gin.Context) {
		c.String(200, "pong")
	})


	r.GET("/testCapture", func(c *gin.Context) {
		result, location := loginAndCapture("10.19.138.110",8000,"admin","abc12345",0,0)
		if result == 0 {
			c.JSON(http.StatusOK, gin.H{"status": result,"cephLocation": location})
		}else {
			c.JSON(http.StatusBadRequest, gin.H{"status": result})
		}
	})

	r.GET("/cameraLogin", func(c *gin.Context) {
		var cameraLogin CameraLogin

		if c.BindJSON(&cameraLogin) == nil {
			port,_ := strconv.Atoi(cameraLogin.Port)
			var result, uid = HikLogin(
				cameraLogin.Address,
				port,
				cameraLogin.UserName,
				cameraLogin.PassWord,
			)
			if result == 0 {

				c.JSON(http.StatusOK, gin.H{"registerCode": 0, "lUserID": uid})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"registerCode": result, "lUserID": 0})
			}
		}
	})

	r.GET("/cameraLogout",func(c *gin.Context){
		var cameraLogout CameraLogout

		if c.Bind(&cameraLogout) == nil{
			uid,_ := strconv.ParseInt(cameraLogout.LUid,10,64)
			HikLogout(uid)
			c.JSON(http.StatusOK, gin.H{"registerCode": 0})
		}
	})

	r.GET("/ConnectionTest", func(c *gin.Context) {
		var cameraLogin CameraLogin

		if c.BindJSON(&cameraLogin) == nil {
			port,_ := strconv.Atoi(cameraLogin.Port)
			var result, uid = ConnectionTest(
				cameraLogin.Address,
				port,
				cameraLogin.UserName,
				cameraLogin.PassWord,
			)
			if result == 0 {

				c.JSON(http.StatusOK, gin.H{"registerCode": 0, "lUserID": uid})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"registerCode": result, "lUserID": 0})
			}
		}
	})

	r.POST("/cameraCapture", func(c *gin.Context) {
		var cameraCapture CameraCapture

		if c.BindJSON(&cameraCapture) == nil {
			port,_ := strconv.Atoi(cameraCapture.Port)
			quality,_ := strconv.Atoi(cameraCapture.Quality)
			size, _ := strconv.Atoi(cameraCapture.Size)
			result, location := loginAndCapture(
				cameraCapture.Address,
				port,
				cameraCapture.UserName,
				cameraCapture.PassWord,
				quality,
				size,
			)
			if result == 0 {
				location = Ceph.Bucket + "/" + location
				c.JSON(http.StatusOK, gin.H{"captureCode": 0, "cephLocation": location})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"captureCode": result, "cephLocation": location})
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

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}