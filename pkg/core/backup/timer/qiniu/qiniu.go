// Package qiniu provides ...
package qiniu

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

// RestoreData implements timer.Storage
func (s Storage) RestoreData() error {
	switch config.Conf.Database.Driver {
	case "mongodb":
		return restoreToMongoDB()
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
	arg = fmt.Sprintf("tar czf /tmp/%s -C /tmp eiblog", name)
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
		Name:           filepath.Join("blog", name), // blog/eiblog-xx.tar.gz
		Size:           s.Size(),
		Data:           f,
		NoCompletePath: true,

		Conf: config.Conf.BackupApp.Qiniu,
	}
	_, err = internal.QiniuUpload(uploadParams)
	if err != nil {
		return err
	}
	// after days delete
	deleteParams := internal.DeleteParams{
		Name:           name,
		Days:           config.Conf.BackupApp.Validity,
		NoCompletePath: true,

		Conf: config.Conf.BackupApp.Qiniu,
	}
	return internal.QiniuDelete(deleteParams)
}

func restoreToMongoDB() error {
	params := internal.ContentParams{
		Prefix: "eiblog",

		Conf: config.Conf.BackupApp.Qiniu,
	}
	raw, err := internal.QiniuContent(params)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("/tmp/eiblog.tar.gz", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, _ = f.Write(raw)
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()
	// unarchive
	arg := fmt.Sprintf("tar xzf /tmp/eiblog.tar.gz -C /tmp")
	cmd := exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return err
	}
	// restore
	arg = fmt.Sprintf("mongorestore -h %s -d eiblog /tmp/eiblog", config.Conf.Database.Source)
	cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	return cmd.Run()
}
