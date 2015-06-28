package fswatcher

import (
	log "github.com/cihub/seelog"
	"gopkg.in/fsnotify.v1"
	"github.com/Nitecon/gosync/datastore"
	"github.com/Nitecon/gosync/storage"
	"github.com/Nitecon/gosync/utils"
	"os"
	"strconv"
)

func SysPathWatcher(path string) {
	log.Infof("Starting new watcher for %s:", path)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Criticalf("Cannot create new watcher for %s\n %s", path, err.Error())
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		listener := utils.GetListenerFromDir(path)
		rel_path := utils.GetRelativePath(listener, path)
		for {
			select {
			case event := <-watcher.Events:
				//logs.WriteLn("event:", event)
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
                    if !utils.MatchesIgnore(path, event.Name){
                        runFileChmod(path, event.Name)
                    }
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {

                    if !utils.MatchesIgnore(path, event.Name){
                        if checksumItem(path, rel_path, event.Name) {
                            log.Infof("Rename occurred on:", event.Name)
                            runFileRename(path, event.Name)
                        }
                    }

				}
				if (event.Op&fsnotify.Create == fsnotify.Create) || (event.Op&fsnotify.Write == fsnotify.Write) {

					if checksumItem(path, rel_path, event.Name) {
                        if !utils.MatchesIgnore(path, event.Name){
                            log.Infof("New / Modified File: %s", event.Name)
                            runFileCreateUpdate(path, event.Name, "create")
                        }

					}
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
                    if !utils.MatchesIgnore(path, event.Name){
                        runFileRemove(path, event.Name)
                        log.Infof("Removed File: %s", event.Name)
                    }

				}
			case err := <-watcher.Errors:
				log.Errorf("error:", err)
			}
		}

	}()
	err = watcher.Add(path)
    if err != nil{
        log.Criticalf("Cannot add watcher to %s\n %s", path, err.Error())
    }
	<-done

}

func runFileRename(base_path, path string) bool {
	log.Infof("File removed %s: ", path)

	return true
}

func runFileRemove(base_path, path string) bool {
	listener := utils.GetListenerFromDir(base_path)
	rel_path := utils.GetRelativePath(listener, path)
	// Remove the item from the database (it's already removed from filesystem
	datastore.Remove(listener, rel_path)
	// Remove the item from the backup storage location
	storage.RemoveFile(rel_path, listener)
	return true
}

func runFileChmod(base_path, path string) bool {
	listener := utils.GetListenerFromDir(base_path)
	rel_path := utils.GetRelativePath(listener, path)
	dbItem, err := datastore.GetOne(base_path, rel_path)
	if err != nil {
		log.Errorf("Error occurred trying to get %s from DB\nError: %s", rel_path, err.Error())
		return false
	}
	fsItem, err := utils.GetFileInfo(path)
	if err != nil {
		log.Errorf("Could not find item on filesystem: %s\nError:%s", path, err.Error())
	}
	if dbItem.Perms != fsItem.Perms {
		iPerm, _ := strconv.Atoi(dbItem.Perms)
		mode := int(iPerm)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Infof("File no longer exists returning")
			return true
		} else {
			err := os.Chmod(path, os.FileMode(mode))
			if err != nil {
				log.Errorf("Error occurred changing file modes: %s", err.Error())
				return false
			}
			return true
		}
	} else {
		log.Info("File modes are correct changing nothing")
		return false
	}
	return true
}

func runFileCreateUpdate(base_path, path, operation string) bool {
	listener := utils.GetListenerFromDir(base_path)
	rel_path := utils.GetRelativePath(listener, path)
	fsItem, err := utils.GetFileInfo(path)

	if err != nil {
		log.Infof("Error getting file details for %s: %+v", path, err)
	}

	log.Infof("Putting in storage:-> %s", rel_path)
	storage.PutFile(path, listener)
	log.Infof("Creating/Updating:-> %s", rel_path)
	return datastore.Insert(listener, fsItem)
}

func checksumItem(base_path, rel_path, abspath string) bool {
	dbItem, err := datastore.GetOne(base_path, rel_path)
	if err != nil {
		log.Infof("Error occurred getting item from DB: %s", err.Error())
		return false
	} else {
		if dbItem.Checksum != utils.GetMd5Checksum(abspath) {
			log.Infof("%s != DB item checksum", abspath)
			return true
		} else {
			log.Infof("%s === DB item checksum", abspath)
			return false
		}
	}
	return false
}
