// Package timer provides ...
package timer

import (
	"errors"
	"strconv"
	"time"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/core/backup/timer/qiniu"

	"github.com/sirupsen/logrus"
)

// Start to backup with ticker
func Start(restore bool) (err error) {
	var storage Storage
	// backup instance
	switch config.Conf.BackupApp.BackupTo {
	case "qiniu":
		storage = qiniu.Storage{}

	default:
		return errors.New("timer: unknown backup to driver: " +
			config.Conf.BackupApp.BackupTo)
	}
	if restore {
		err = storage.RestoreData()
		if err != nil {
			return err
		}
		logrus.Info("timer: RestoreData success")
	}
	// parse duration
	interval, err := ParseDuration(config.Conf.BackupApp.Interval)
	if err != nil {
		return err
	}
	t := time.NewTicker(interval)
	for now := range t.C {
		err = storage.BackupData(now)
		if err != nil {
			logrus.Error("timer: Start.BackupData: ", now, err)
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

// Storage backup backend
type Storage interface {
	BackupData(now time.Time) error
	RestoreData() error
}
