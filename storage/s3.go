package storage

import (
	"github.com/rlmcpherson/s3gof3r"
    "github.com/Nitecon/gosync/utils"
	"os"
    "io"
    "github.com/Nitecon/gosync/datastore"
    "strings"
    log "github.com/cihub/seelog"
)

type S3 struct {
	config   *utils.Configuration
	bucket   *s3gof3r.Bucket
	listener string
}

type Keys struct {
    AccessKey     string
    SecretKey     string
    SecurityToken string

}

func (s *S3) Upload(local_path, remote_path string) error {
    conf,keys := s.GetS3Config()

    // Open bucket to put file into
    s3 := s3gof3r.New("", *keys)
    b := s3.Bucket(s.config.Listeners[s.listener].Bucket)

    // open file to upload
    file, err := os.Open(local_path)
    if err != nil {
        return err
    }

    // Open a PutWriter for upload
    w, err := b.PutWriter(remote_path, nil, conf)
    if err != nil {
        return err
    }
    if _, err = io.Copy(w, file); err != nil { // Copy into S3
        return err
    }
    if err = w.Close(); err != nil {
        return err
    }
    return nil

}

func (s *S3) Download(remote_path, local_path string,uid, gid int, perms, listener string) error {
    log.Infof("S3 Downloading %s", local_path)
    conf,keys := s.GetS3Config()

    // Open bucket to put file into
    s3 := s3gof3r.New("", *keys)
    b := s3.Bucket(s.config.Listeners[s.listener].Bucket)

    r, _, err := b.GetReader(remote_path, conf)
    if err != nil {
        return err
    }
    // stream to file
    if _, err = utils.FileWrite(local_path, r, uid,gid, perms); err != nil {
        return err
    }
    err = r.Close()
    if err != nil {
        return err
    }
    basePath := s.config.Listeners[listener].BasePath
    datastore.UpdateHost(listener, strings.TrimLeft(local_path, basePath))

    return err
}

func (s *S3) Remove(remote_path string) bool {
    log.Infof("Removing file %s from s3 storage", remote_path)
    _,keys := s.GetS3Config()

    // Open bucket to put file into
    s3 := s3gof3r.New("", *keys)
    b := s3.Bucket(s.config.Listeners[s.listener].Bucket)
    b.Delete(remote_path)
	return true
}

func (s *S3) GetS3Config() (*s3gof3r.Config, *s3gof3r.Keys){
    conf := new(s3gof3r.Config)
    *conf = *s3gof3r.DefaultConfig
    keys := new(s3gof3r.Keys)
    keys.AccessKey = s.config.S3Config.Key
    keys.SecretKey = s.config.S3Config.Secret
    conf.Concurrency = 10
    return conf, keys
}

func getListener(dir string) string {
	cfg := utils.GetConfig()
	var listener = ""
	for lname, ldata := range cfg.Listeners {
		if ldata.Directory == dir {
			listener = lname
		}
	}
	return listener
}
