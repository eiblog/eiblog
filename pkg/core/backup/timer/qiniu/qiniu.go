// Package qiniu provides ...
package qiniu

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	switch config.Conf.Database.Driver {
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
	u, err := url.Parse(config.Conf.Database.Source)
	if err != nil {
		return err
	}
	arg := fmt.Sprintf("mongodump -h %s -d eiblog -o /tmp", u.Host)
	cmd := exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return err
	}
	// tar
	name := fmt.Sprintf("eiblog-%s.tar.gz", now.Format("2006-01-02"))
	arg = fmt.Sprintf("tar czf /tmp/%s /tmp/eiblog", name)
	cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return err
	}

	// upload file
	f, err := os.Open("/tmp/" + name)
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
		Days: config.Conf.BackupApp.Validity,

		Conf: config.Conf.BackupApp.Qiniu,
	}
	return internal.QiniuDelete(deleteParams)
}
