package to

// BackupRestorer 备份存储接口
type BackupRestorer interface {
	Upload(path string) error
	Download() (path string, err error)
}
