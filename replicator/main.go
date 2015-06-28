package replicator

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/Nitecon/gosync/datastore"
	"github.com/Nitecon/gosync/nodeinfo"
	"github.com/Nitecon/gosync/storage"
	"github.com/Nitecon/gosync/utils"
	"os"
	"time"
)

func getFileInDatabase(dbPath string, fsItems []utils.FsItem) bool {
	for _, fsitem := range fsItems {
		if fsitem.Filename == dbPath {
			return true
		}
	}
	return false
}

func InitialSync() {
	cfg := utils.GetConfig()
	log.Info("Verifying DB Tables")
	datastore.CreateDB()
	log.Info("Initial sync starting...")

	for key, listener := range cfg.Listeners {

		// First check to see if the table is empty and do a full import false == not empty
		if datastore.CheckEmpty(key) == false {
			// Database is not empty so pull the updates and match locally
			items := datastore.FetchAll(key)
			handleDataChanges(items, listener, key)
		} else {
			// Database is empty so lets import
			fsItems := utils.ListFilesInDir(listener.Directory)
			for _, item := range fsItems {
				success := datastore.Insert(key, item)
				if success != true {
					log.Infof("An error occurred inserting %x to database", item)
				}
				if !item.IsDir {
					storage.PutFile(item.Filename, key)
				}

			}
		}

	}

	log.Info("Initial sync completed...")
}

func CheckIn(path string) {
	log.Info("Starting db checking background script: " + path)
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				nodeinfo.UpdateConnection()
				log.Info("Checking all changed stuff in db for: " + path)
				listener := utils.GetListenerFromDir(path)
				items, err := datastore.CheckIn(listener)
				if err != nil {
					log.Infof("Error occurred getting data for %s (%s): %+v", listener, err.Error(), err)
				}
				cfg := utils.GetConfig()
				handleDataChanges(items, cfg.Listeners[listener], listener)
			// @TODO: check that db knows Im alive.
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func handleDataChanges(items []utils.DataTable, listener utils.Listener, listenerName string) {
	// Walking the directory to get files.
	fsItems := utils.ListFilesInDir(listener.Directory)
	itemsToDelete := findOnlyInFS(fsItems, items, listenerName)
	if len(itemsToDelete) > 0 {
		for _, delItem := range itemsToDelete {
			if !utils.MatchesIgnore(listener.Directory, delItem) {
				itemExists := datastore.PathExists(listenerName, delItem)
				if !itemExists {
					log.Infof("Removing untracked item: %s", delItem)
					err := os.Remove(delItem)
					if err != nil {
						log.Infof("Error deleting untracked file %s\n%s", delItem, err.Error())
					}
				}
			}

		}
	}
	for _, item := range items {
		absPath := utils.GetAbsPath(listenerName, item.Path)
		itemMatch := getFileInDatabase(absPath, fsItems)
		perms := fmt.Sprintf("File Permissions: %#o", item.Perms)
		if itemMatch {
			// Check to make sure it's not a directory as directories don't need to be uploaded
			if !item.IsDirectory {
				fileMD5 := utils.GetMd5Checksum(absPath)
				if fileMD5 != item.Checksum {
					log.Info("Processing modified file: " + absPath)
					hostname, _ := os.Hostname()
					if item.HostUpdated != hostname {
						// File wasn't updated locally so get it from remote
						bNodeCopy := storage.GetNodeCopy(item, listenerName, listener.Uid, listener.Gid, perms)
						if bNodeCopy {
							// Item downloaded successfully so update db with new md5
							newFileInfo, err := utils.GetFileInfo(absPath)
							if err != nil {
								log.Infof("Error occurred trying to get info on file: %s", absPath)
							} else {
								datastore.Insert(listenerName, newFileInfo)
							}
						} else {
							// Item didn't download from nodes (maybe ports are closed so lets get from backups)
							storage.GetFile(absPath, listenerName, listener.Uid, listener.Gid, perms)
							newFileInfo, err := utils.GetFileInfo(absPath)
							if err != nil {
								log.Infof("Error occurred trying to get info on file: %s", absPath)
							} else {
								datastore.Insert(listenerName, newFileInfo)
							}
						}
					} else {
						// Item was modified locally and does not match so lets fix it with what is in db
						storage.GetFile(absPath, listenerName, listener.Uid, listener.Gid, perms)
						newFileInfo, err := utils.GetFileInfo(absPath)
						if err != nil {
							log.Infof("Error occurred trying to get info on file: %s", absPath)
						} else {
							datastore.Insert(listenerName, newFileInfo)
						}
					}
				}
			}
			// Now we check to make sure the files match correct users etc

		} else {
			if !utils.MatchesIgnore(listener.Directory, absPath) {
				// Item doesn't exist locally but exists in DB so restore it
				log.Infof("Item Deleted Locally: %s restoring from DB marker", absPath)
				if item.IsDirectory {
					dirExists, _ := utils.ItemExists(absPath)
					if !dirExists {
						os.MkdirAll(absPath, 0775)
					}
				} else {
					if !storage.GetNodeCopy(item, listenerName, listener.Uid, listener.Gid, perms) {
						log.Infof("Server is down for %s going to backup storage", listenerName)
						// The server must be down so lets get it from S3
						storage.GetFile(absPath, listenerName, listener.Uid, listener.Gid, perms)
					}
				}
			}

		}
	}
}

func findOnlyInFS(fsItems []utils.FsItem, dbItems []utils.DataTable, listener string) []string {
	leftovers := []string{}
	for _, fsItem := range fsItems {
		itemInDB := false
		// Iterate over db items to check and see if the item exists in the db
		for _, dbItem := range dbItems {
			absPath := utils.GetAbsPath(listener, dbItem.Path)
			//log.Infof("Validating if %s == %s", fsItem.Filename, absPath)
			if fsItem.Filename == absPath {
				itemInDB = true
			}
		}
		// Item doesn't exist in db so add it to the leftover array
		if !itemInDB {
			leftovers = append(leftovers, fsItem.Filename)
		}
	}
	return leftovers
}
