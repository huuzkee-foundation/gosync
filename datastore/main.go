package datastore

import (
	"github.com/Nitecon/gosync/utils"
    log "github.com/cihub/seelog"
)

var (
	dbstore Datastore
)

type Datastore interface {
	Insert(table string, item utils.FsItem) bool
    Remove(table, path string) bool
	CheckEmpty(table string) bool
	FetchAll(table string) []utils.DataTable
	CheckIn(listener string) ([]utils.DataTable, error)
    GetOne(listener, path string) (utils.DataTable, error)
    UpdateHost(table, path string)
    PathExists(listener, path string) bool
	CreateDB()
}

func setdbstoreEngine() {
	cfg := utils.GetConfig()
	var engine = cfg.Database.Type
	switch engine {
	case "mysql":
		dbstore = &MySQLDB{config: cfg}
	}
}

func Insert(table string, item utils.FsItem) bool {
	setdbstoreEngine()
	return dbstore.Insert(table, item)
}

func CheckEmpty(table string) bool {
	setdbstoreEngine()
	empty := dbstore.CheckEmpty(table)
	if empty {
        log.Info("Database is EMPTY, starting creation")
	} else {
        log.Info("Using existing table: " + table)
	}
	return empty
}

func FetchAll(table string) []utils.DataTable {
	setdbstoreEngine()
	return dbstore.FetchAll(table)
}

func UpdateHost(table, path string){
    setdbstoreEngine()
    dbstore.UpdateHost(table, path)
}

func CheckIn(listener string) ([]utils.DataTable, error) {
    setdbstoreEngine()
    log.Infof("Starting db checking background script for: " + listener)
	data,err := dbstore.CheckIn(listener)
    return data, err

}

func GetOne(basepath, path string) (utils.DataTable, error){
    setdbstoreEngine()
    listener := utils.GetListenerFromDir(basepath)
    dbitem, err := dbstore.GetOne(listener, path)
    return dbitem, err
}

func PathExists(listener, path string) bool{
    setdbstoreEngine()
    return dbstore.PathExists(listener, utils.GetRelativePath(listener, path))
}

func Remove(table, path string) bool {
    setdbstoreEngine()
    return dbstore.Remove(table, path)
}

func CreateDB() {
    setdbstoreEngine()
	dbstore.CreateDB()
}
