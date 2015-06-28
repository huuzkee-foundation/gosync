package nodeinfo

import (
    "github.com/Nitecon/gosync/utils"
)

var (
    nodeStore NodeStore
)

type NodeStore interface {
    Insert() bool
    GetAll() []utils.NodeTable
    CreateDB()
    Update() bool
    Close(deadHost string) // call this method when you want to close the connection
    initDB()
}



func setdbstoreEngine() {
    cfg := utils.GetConfig()
    var engine = cfg.Database.Type
    switch engine {
        case "mysql":
        nodeStore = &MySQLNodeDB{config: cfg}
        nodeStore.initDB()
    }
}

func Initialize(){
    setdbstoreEngine()
    nodeStore.CreateDB()
    nodeStore.initDB()
}

func SetAlive() bool {
    setdbstoreEngine()
    return nodeStore.Insert()
}

func UpdateConnection() bool {
    setdbstoreEngine()
    return nodeStore.Update()
}

func CloseConnection(deadHost string) {
    setdbstoreEngine()
    nodeStore.Close(deadHost)
}

func GetNodes() []utils.NodeTable{
    setdbstoreEngine()
    return nodeStore.GetAll()
}