package storage

import (
    "github.com/Nitecon/gosync/utils"
    log "github.com/cihub/seelog"
)

type GDrive struct {
	config *utils.Configuration
    listener string
}

func (g *GDrive) Upload(local_path, remote_path string) error {
    log.Info("Doing an upload to GDrive")
	return nil
}

func (g *GDrive) Download(remote_path, local_path string, uid, gid int, perms, listener string) error {
    log.Info("Doing a download from GDrive")
	return nil
}

func (g *GDrive) Remove(remote_path string) bool {
    log.Infof("Removing %s from Google Drive Storage", remote_path)
	return true
}
