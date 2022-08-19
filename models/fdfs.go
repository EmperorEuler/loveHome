package models

import (
	"loveHome/models"

	"github.com/astaxie/beego"
	"github.com/keonjeo/fdfs_client"
)

// 上传单个文件到fdfs函数，设置3个返回值，组名，文件id，错误
func TestUploadByFilename(fileName string) (groupName string, FileId string, err error) {
	//连接fdfs通过client.conf 文件
	fdfsClient, errClient := fdfs_client.NewFdfsClient("../conf/client.conf")
	//连不上就报错
	if err != nil {
		beego.Info("Newo FdfsClient error  %s", errClient.Error())
		return "", "", errClient
	}
	//连上后，上传文件，返回组名，文件id
	uploadResponse, errUpload := fdfsClient.UploadByFilename(fileName)
	if errUpload != nil {
		beego.Info("New FdfsClient err %s", errUpload.Error())
		return "", "", errUpload
	}
	return uploadResponse.GroupName, uploadResponse.RemoteFileId, nil
}

// beego有getfile函数直接能得到二进制byte 就不需要跟原作者一样慢慢获取 直接封装就可以调用
func UploadByBuffer(fileBuffer []byte, suffixStr string) (uploadResp *fdfs_client.UploadFileResponse, err error) {
	resp := make(map[string]interface{})
	//连接到fdfs服务器
	fdfsClient, err := fdfs_client.NewFdfsClient("conf/client.conf")
	if err != nil {
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return nil, err
	}

	//把文件上传到fdfs上
	uploadResponse, err := fdfsClient.UploadByBuffer(fileBuffer, suffixStr)

	if err != nil {
		beego.Error("testuploadbybuffer error %s", err.Error())
		resp["errno"] = models.RECODE_DBERR
		resp["errmsg"] = models.RecodeText(models.RECODE_DBERR)
		return nil, err
	}

	beego.Info(uploadResponse.GroupName)
	beego.Info(uploadResponse.RemoteFileId)
	return uploadResponse, nil
}
