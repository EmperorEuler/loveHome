package models

import (
	"github.com/keonjeo//fdfs_client"
	"github.com/astaxie/beego"
)

func TestUploadByFilename(fileName string) (groupName string, FileId string, err error) {
	fdfsClient, errClient := fdfs_client.NewFdfsClient("../conf/client.conf")
	if err != nil {
		beego.Info("Newo FdfsClient error  %s", errClient.Error())
		return "", "", errClient
	}
	uploadResponse, errUpload := fdfsClient.UploadByFilename("fileName")
	if errUpload != nil {
		beego.Info("New FdfsClient err %s", errUpload.Error())
		return "", "", errUpload
	}
	return uploadResponse.GroupName, uploadResponse.RemoteFileId, nil
}
