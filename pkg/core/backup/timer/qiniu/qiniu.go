// Package qiniu provides ...
package qiniu

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/internal"
)

// Storage qiniu storage
type Storage struct{}

// BackupData implements timer.Storage
func (s Storage) BackupData(now time.Time) error {
	switch config.Conf.Database.Source {
	case "mongodb":
		return backupFromMongoDB(now)
	default:
		return errors.New("unsupported database source backup to qiniu")
	}
}

func backupFromMongoDB(now time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()

	// dump
	arg := fmt.Sprintf("mongodump -h %s -d eiblog -o /tmp",
		config.Conf.Database.Source)
	cmd := exec.CommandContext(ctx, "sh", "-c", arg)
	err := cmd.Run()
	if err != nil {
		return err
	}
	// tar
	name := fmt.Sprintf("eiblog-%s.tar.gz", now.Format("2006-01-02"))
	arg = fmt.Sprintf("tar czf %s /tmp/eiblog", name)
	cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return err
	}

	// upload file
	f, err := os.Open("/tmp/eiblog/" + name)
	if err != nil {
		return err
	}
	s, err := f.Stat()
	if err != nil {
		return err
	}
	uploadParams := internal.UploadParams{
		Name: name,
		Size: s.Size(),
		Data: f,

		Conf: config.Conf.BackupApp.Qiniu,
	}
	_, err = internal.QiniuUpload(uploadParams)
	if err != nil {
		return err
	}
	// after days delete
	deleteParams := internal.DeleteParams{
		Name: name,

		Conf: config.Conf.BackupApp.Qiniu,
	}
	return internal.QiniuDelete(deleteParams)
}
