package timer

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/eiblog/eiblog/cmd/backup/config"
	"github.com/eiblog/eiblog/cmd/backup/handler/timer/db"
	"github.com/eiblog/eiblog/cmd/backup/handler/timer/to"

	"github.com/sirupsen/logrus"
)

// Start to backup with ticker
func Start(restore bool) (err error) {
	var (
		storage  db.Storage
		backupTo to.BackupRestorer
	)

	// backup from
	switch config.Conf.Database.Driver {
	case "mongodb":
		storage = db.MongoStorage{}

	default:
		return errors.New("timer: unknown backup from driver: " +
			config.Conf.Database.Driver)
	}

	// backup to
	switch config.Conf.BackupTo {
	case "qiniu":
		backupTo = to.QiniuBackupRestorer{}

	default:
		return errors.New("timer: unknown backup to driver: " +
			config.Conf.BackupTo)
	}

	// restore
	if restore {
		path, err := backupTo.Download()
		if err != nil {
			return err
		}
		err = storage.Restore(path)
		if err != nil {
			return err
		}
		logrus.Info("timer: Restore success")
	}

	// backup
	interval, err := ParseDuration(config.Conf.Interval)
	if err != nil {
		return err
	}
	t := time.NewTicker(interval)
	for now := range t.C {
		name := fmt.Sprintf("eiblog-%s.tar.gz", now.Format("2006-01-02"))
		path, err := storage.Backup(name)
		if err != nil {
			logrus.Error("timer: Start.Backup: ", now.Format(time.RFC3339), err)
			continue
		}
		err = backupTo.Upload(path)
		if err != nil {
			logrus.Error("timer: Start.Backup: ", now.Format(time.RFC3339), err)
		}
	}
	return nil
}

// ParseDuration parse string to duration
func ParseDuration(d string) (time.Duration, error) {
	if len(d) == 0 {
		return 0, errors.New("timer: incorrect duration input")
	}

	length := len(d)
	switch d[length-1] {
	case 's', 'm', 'h':
		return time.ParseDuration(d)
	case 'd':
		di, err := strconv.Atoi(d[:length-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(di) * time.Hour * 24, nil
	}
	return 0, errors.New("timer: unsupported duration:" + d)
}
