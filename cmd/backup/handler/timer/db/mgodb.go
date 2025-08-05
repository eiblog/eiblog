package db

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"time"

	"github.com/eiblog/eiblog/cmd/backup/config"
	pdb "github.com/eiblog/eiblog/pkg/connector/db"
)

// MongoStorage 备份恢复器
type MongoStorage struct{}

// Backup 备份
func (r MongoStorage) Backup(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// dump
	u, err := url.Parse(config.Conf.Database.Source)
	if err != nil {
		return "", err
	}
	if u.Path == "" {
		return "", fmt.Errorf("no database specified")
	}
	arg := fmt.Sprintf("mongodump -h %s -d %s -o /tmp", u.Host, u.Path)
	cmd := exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	// tar
	arg = fmt.Sprintf("tar czf /tmp/%s -C /tmp %s", name, u.Path)
	cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return "/tmp/" + name, nil
}

// Restore 恢复
func (r MongoStorage) Restore(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// drop database
	database, err := pdb.NewMDB(ctx, config.Conf.Database)
	if err != nil {
		return err
	}
	err = database.Drop(ctx)
	if err != nil {
		return err
	}
	// unarchive
	arg := fmt.Sprintf("tar xzf %s -C /tmp", path)
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
	arg = fmt.Sprintf("mongorestore -h %s -d %s /tmp/%s", u.Host, u.Path, u.Path)
	cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	return cmd.Run()
}
