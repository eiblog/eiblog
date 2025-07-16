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

	"github.com/eiblog/eiblog/cmd/backup/config"
	"github.com/eiblog/eiblog/cmd/backup/handler/internal"
	"github.com/eiblog/eiblog/pkg/connector/db"
	"github.com/eiblog/eiblog/pkg/third/qiniu"
)

// BackupRestorer qiniu backup restorer
type BackupRestorer struct{}

// Backup implements timer.BackupRestorer
func (s BackupRestorer) Backup(now time.Time) error {
	switch config.Conf.Database.Driver {
	case "mongodb":
		return backupFromMongoDB(now)
	default:
		return errors.New("unsupported source backup to qiniu: " +
			config.Conf.Database.Driver)
	}
}

// Restore implements timer.BackupRestorer
func (s BackupRestorer) Restore() error {
	switch config.Conf.Database.Driver {
	case "mongodb":
		return restoreToMongoDB()
	default:
		return errors.New("unsupported source restore from qiniu: " +
			config.Conf.Database.Driver)
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
	uploadParams := qiniu.UploadParams{
		Name:           filepath.Join("blog", name), // blog/eiblog-xx.tar.gz
		Size:           s.Size(),
		Data:           f,
		NoCompletePath: true,
	}
	_, err = internal.QiniuClient.Upload(uploadParams)
	if err != nil {
		return err
	}
	// after days delete
	deleteParams := qiniu.DeleteParams{
		Name:           filepath.Join("blog", name), // blog/eiblog-xx.tar.gz
		Days:           config.Conf.Validity,
		NoCompletePath: true,
	}
	return internal.QiniuClient.Delete(deleteParams)
}

func restoreToMongoDB() error {
	// backup file
	params := qiniu.ContentParams{
		Prefix: "blog/",
	}
	raw, err := internal.QiniuClient.Content(params)
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
	// drop database
	mdb, err := db.NewMDB(config.Conf.Database)
	if err != nil {
		return err
	}
	err = mdb.Drop(ctx)
	if err != nil {
		return err
	}
	// unarchive
	arg := "tar xzf /tmp/eiblog.tar.gz -C /tmp"
	cmd := exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return err
	}
	// restore
	u, err := url.Parse(config.Conf.Database.Source)
	if err != nil {
		return err
	}
	arg = fmt.Sprintf("mongorestore -h %s -d eiblog /tmp/eiblog", u.Host)
	cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	return cmd.Run()
}
