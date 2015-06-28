package utils

import (
	"time"
)

type DataTable struct {
	Id          int       `db:"id"`
	Path        string    `db:"path"`
	IsDirectory bool      `db:"is_dir"`
	Filename    string    `db:"filename"`
	Directory   string    `db:"directory"`
	Checksum    string    `db:"checksum"`
	Atime       time.Time `db:"atime"`
	Mtime       time.Time `db:"mtime"`
	Perms       string    `db:"perms"`
	HostUpdated string    `db:"host_updated"`
	HostIPs     string    `db:"host_ips"`
	LastUpdate  time.Time `db:"last_update"`
}

type NodeTable struct {
	Id         int       `db:"id"`
	HostName   string    `db:"hostname"`
	NodeIPs    string    `db:"host_ips"`
	Connected  time.Time `db:"connected_on"`
	LastUpdate time.Time `db:"last_update"`
}

type FsItem struct {
	Filename string
	Path     string
	Checksum string
	IsDir    bool
	Mtime    time.Time
	Perms    string
}

type AppError struct {
	Error   error
	Message string
	Code    int
	Stack   string
}

type Configuration struct {
	ServerConfig BaseConfig
	S3Config     StorageS3 `toml:"StorageS3"`
	GDConfig     StorageGDrive
	Database     Database
	Listeners    map[string]Listener
}

type Database struct {
	Type string
	Dsn  string
}

type StorageS3 struct {
	Key    string `toml:"key"`
	Secret string `toml:"secret"`
	Region string `toml:"region"`
}

type StorageGDrive struct {
	Key    string
	Secret string
	Region string
}

type BaseConfig struct {
	ListenPort  string `toml:"listen_port"`
	RescanTime  int    `toml:"rescan"`
	StorageType string `toml:"storagetype"`
	LogLocation string `toml:"log_location"`
	LogLevel    string    `toml:"log_level"`
}

type Listener struct {
	Directory   string
	Excludes    string `toml:"excludes"`
	Uid         int
	Gid         int
	StorageType string `toml:"storagetype"`
	Bucket      string `toml:"bucket"`
	BasePath    string `toml:"basepath"`
}
