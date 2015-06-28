package storage

import (

    "github.com/Nitecon/gosync/utils"
    log "github.com/cihub/seelog"

)

var (
	// BackupConfig determines which storage system to use,
	// and we then assign that to "storage".
	storage Backupper
)

type Backupper interface {
	Upload(local_path, remote_path string) error
	Download(remote_path, local_path string, uid, gid int, perms, listener string) error
	Remove(remote_path string) bool
}

func setStorageEngine(listener string) {
	cfg := utils.GetConfig()
	var engine = cfg.Listeners[listener].StorageType
	switch engine {
	case "gdrive":
		storage = &GDrive{config:cfg, listener:listener}
	case "s3":
		storage = &S3{config:cfg,listener:listener}
	}
}

func PutFile(local_path, listener string) error {
	setStorageEngine(listener)
	return storage.Upload(local_path, utils.GetRelativeUploadPath(listener, local_path))
}

func GetFile(local_path, listener string, uid, gid int, perms string) error {
	setStorageEngine(listener)
	err := storage.Download(utils.GetRelativeUploadPath(listener, local_path), local_path, uid, gid, perms, listener)
    if err != nil{
        log.Infof("Error downloading file from S3 (%s) : %+v", err.Error(), err)
    }
    return err
}

func RemoveFile(local_path, listener string) bool {
	setStorageEngine(listener)
	return storage.Remove(utils.GetRelativeUploadPath(listener, local_path))
}
