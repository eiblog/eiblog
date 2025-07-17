package db

// Storage 备份恢复器
type Storage interface {
	Backup(name string) (string, error)
	Restore(path string) error
}
