package timer

import (
	"errors"
	"strconv"
	"time"

	"github.com/eiblog/eiblog/cmd/backup/config"
	"github.com/eiblog/eiblog/cmd/backup/handler/timer/qiniu"

	"github.com/sirupsen/logrus"
)

// BackupRestorer 备份恢复器
type BackupRestorer interface {
	Backup(now time.Time) error
	Restore() error
}

// Start to backup with ticker
func Start(restore bool) (err error) {
	var storage BackupRestorer

	// backup instance
	switch config.Conf.BackupTo {
	case "qiniu":
		storage = qiniu.BackupRestorer{}

	default:
		return errors.New("timer: unknown backup to driver: " +
			config.Conf.BackupTo)
	}
	if restore {
		err = storage.Restore()
		if err != nil {
			return err
		}
		logrus.Info("timer: Restore success")
	}
	// parse duration
	interval, err := ParseDuration(config.Conf.Interval)
	if err != nil {
		return err
	}
	t := time.NewTicker(interval)
	for now := range t.C {
		err = storage.Backup(now)
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
