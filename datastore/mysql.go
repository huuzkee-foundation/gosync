package datastore

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/Nitecon/gosync/utils"
	"os"
	"path"
	"time"
    log "github.com/cihub/seelog"

    "html"
)

type MySQLDB struct {
	config *utils.Configuration
	db     *sqlx.DB
}

func (my *MySQLDB) Insert(table string, item utils.FsItem) bool {
    my.db = my.initDB()
    defer my.db.Close()
	var isDirectory = 0
	if item.IsDir {
		isDirectory = 1
	}
	hostname, _ := os.Hostname()
	var keyExists = []utils.DataTable{}
    relFileName := string(utils.GetRelativePath(table, item.Filename))
    escapedPath := html.EscapeString(relFileName)
	var query = fmt.Sprintf("SELECT id, path FROM %s WHERE path='%s' LIMIT 1", table, escapedPath)
	err := my.db.Select(&keyExists, query)
	if err != nil {
        log.Infof("Error checking for existence of key: %s, in table %s\n %+v", utils.GetRelativePath(table, escapedPath), table, err)
	}
    newChecksum := utils.GetMd5Checksum(item.Filename)

	tx := my.db.MustBegin()
    log.Infof("Existing items in DB: %d", len(keyExists))
	if len(keyExists) > 0 {
		rowId := fmt.Sprintf("%d", keyExists[0].Id)
		tx.MustExec("UPDATE "+table+" SET path=?, is_dir=?, filename=?, directory=?, checksum=?, atime=?, mtime=?, perms=?, host_updated=?, host_ips=?, last_update=? WHERE id='"+rowId+"'",
			utils.GetRelativePath(table, escapedPath),
			isDirectory,
			path.Base(item.Filename),
			path.Dir(item.Filename),
            newChecksum,
			time.Now().UTC(),
			item.Mtime,
			item.Perms,
			hostname,
			utils.GetLocalIp(),
			time.Now().UTC(),
		)
		err = tx.Commit()
        if err != nil{
            log.Infof("Error updating data in DB table %s, for file %s\n%s\n%+v", table, escapedPath, err.Error(),err)
            return false
        }else{
            log.Infof("Data updated in table %s for file %s", table, escapedPath)
            return true
        }
	} else {
		tx.MustExec("INSERT INTO "+table+" (path, is_dir, filename, directory, checksum, atime, mtime, perms, host_updated, host_ips, last_update) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
			utils.GetRelativePath(table, escapedPath),
			isDirectory,
			path.Base(item.Filename),
			path.Dir(item.Filename),
			utils.GetMd5Checksum(item.Filename),
			time.Now().UTC(),
			item.Mtime,
			item.Perms,
			hostname,
            utils.GetLocalIp(),
			time.Now().UTC())
		err = tx.Commit()
        if err != nil{
            log.Infof("Error inserting data to DB table %s, for file %s\n%s\n%+v", table, escapedPath, err.Error(),err)
            return false
        }else{
            log.Infof("Data inserted to table %s for file %s", table, escapedPath)
            return true
        }

	}
	return true
}

func (my *MySQLDB) UpdateHost(table, path string){
    my.db = my.initDB()
    defer my.db.Close()
    hostname, _ := os.Hostname()
    tx := my.db.MustBegin()
    tx.MustExec("UPDATE "+table+" SET host_updated=? WHERE path='"+path+"'", hostname )
    err := tx.Commit()
    if err != nil{
        log.Criticalf("Error updating host for path: %s", err.Error())
    }
}

func (my *MySQLDB) CheckEmpty(table string) bool {
    my.db = my.initDB()
    defer my.db.Close()
	var count int
	err := my.db.Get(&count, "SELECT count(*) FROM "+table+";")
    if err != nil{
        log.Criticalf("Error counting items in table (%s)", err.Error())
    }

	var isEmpty = true
	if count > 0 {
		isEmpty = false
	}
	return isEmpty
}

func (my *MySQLDB) FetchAll(table string) []utils.DataTable {
    my.db = my.initDB()
    defer my.db.Close()
	dTable := []utils.DataTable{}
	query := "SELECT path, is_dir, checksum, mtime, perms, host_updated, host_ips FROM " + table + " ORDER BY is_dir DESC"
	err := my.db.Select(&dTable, query)
    if err != nil{
        log.Criticalf("Could not fetch items from database for %s\n%s", table, err.Error())
    }
	return dTable
}

func (my *MySQLDB) CheckIn(table string) ([]utils.DataTable, error) {
    my.db = my.initDB()
    defer my.db.Close()
	hostname, _ := os.Hostname()
	dTable := []utils.DataTable{}
	query := fmt.Sprintf("SELECT path, is_dir, checksum, mtime, perms, host_updated, host_ips FROM %s where host_updated != '%s' ORDER BY is_dir DESC, last_update DESC", table, hostname)
	err := my.db.Select(&dTable, query)
	return dTable, err

}

func (my *MySQLDB) GetOne(listener, path string) (utils.DataTable, error){
    my.db = my.initDB()
    defer my.db.Close()
    dItem := utils.DataTable{}
    query := fmt.Sprintf("SELECT path, is_dir, checksum, mtime, perms, host_ips FROM %s where path='%s'", listener, path)
    err := my.db.Get(&dItem, query)
    return dItem, err
}

func (my *MySQLDB) PathExists(listener, path string) bool{
    my.db = my.initDB()
    defer my.db.Close()
    dbItem := utils.DataTable{}
    query := fmt.Sprintf("SELECT path FROM %s where path='%s'", listener, path)
    err := my.db.Get(&dbItem, query)
    if err != nil{
        log.Infof("Item does not exist in db or sql error: %s", err.Error())
        return false
    }
    if dbItem.Path == path{
        return true
    }else{
        return false
    }
    return false
}

func (my *MySQLDB) Remove(table, path string) bool{
    my.db = my.initDB()
    defer my.db.Close()
    query := fmt.Sprintf("DELETE FROM %s where path ='%s'", table, path)
    log.Infof("Executing: %s", query)
    tx := my.db.MustBegin()
    tx.MustExec(query)
    err := tx.Commit()
    if err != nil{
        log.Infof("Error occurrred removing %s:\n%s",path, err.Error())
        return false
    }
    return true
}

func (my *MySQLDB) CreateDB() {
    my.db = my.initDB()
    defer my.db.Close()
    log.Info("Database initialized")
	for key, _ := range my.config.Listeners {
		my.db.MustExec(createTableQuery(key))
	}

}

func (my *MySQLDB) initDB() *sqlx.DB{
	tempdb, err := sqlx.Connect("mysql", my.config.Database.Dsn+"&parseTime=True")
	if err != nil {
        log.Info(err.Error())
	}
	return tempdb
}


func createTableQuery(table string) string {
	var createStmt = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	id int(10) unsigned NOT NULL AUTO_INCREMENT,
  path varchar(4096) COLLATE utf8_unicode_ci NOT NULL,
  is_dir int NOT NULL,
  filename varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  directory varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  checksum varchar(2048) COLLATE utf8_unicode_ci NOT NULL,
  atime timestamp NOT NULL,
  mtime timestamp NOT NULL,
  perms varchar(12) COLLATE utf8_unicode_ci NOT NULL,
  host_updated varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  host_ips varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  last_update timestamp NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
`, table)

	return createStmt
}
