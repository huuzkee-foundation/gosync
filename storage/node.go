package storage

import (
	"net/http"
    "github.com/Nitecon/gosync/utils"
    "github.com/Nitecon/gosync/nodeinfo"
    "strings"
    "time"
    log "github.com/cihub/seelog"
)


func GetNodeCopy(item utils.DataTable, listener string, uid, gid int, perms string) bool {
    cfg := utils.GetConfig()
    aliveNodes := nodeinfo.GetNodes()
    for _, node := range aliveNodes{
        log.Infof("Trying download from: %s", node.NodeIPs)
        nIPs := strings.Split(node.NodeIPs, ",")
        for _, ipAddress := range nIPs{
            resp, err := getData(ipAddress, cfg.ServerConfig.ListenPort, listener, utils.GetRelativePath(listener, item.Path))
            if err == nil{
                defer resp.Body.Close()
                if resp.Status == "404" {
                    log.Infof("File not found: %s", item.Path)
                    return false
                }
                size, err := utils.FileWrite(item.Path, resp.Body, uid, gid, perms)
                if err != nil{
                    log.Infof("Cannot write file: %s", item.Path)
                    return false
                }else{
                    log.Infof("%s with %v bytes downloaded", item.Path, size)
                    return true
                }
                return false
            }else{
                return false
            }
        }
    }
    return false
}


func getData(hostname, port, listener, path string) (*http.Response, error){
    rawURL := "http://" + hostname + ":" + port + "/" + listener + path
    log.Infof("Download attempt on: %s", rawURL)
    timeout := time.Duration(5 * time.Second)
    client := http.Client{
        Timeout: timeout,
    }
    resp, err := client.Get(rawURL) // add a filter to check redirect
    return resp, err
}