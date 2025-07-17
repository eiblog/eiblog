package to

import (
	"os"
	"path/filepath"

	"github.com/eiblog/eiblog/cmd/backup/config"
	"github.com/eiblog/eiblog/cmd/backup/handler/internal"
	"github.com/eiblog/eiblog/pkg/third/qiniu"
)

// QiniuBackupRestorer qiniu backup restorer
type QiniuBackupRestorer struct{}

// Upload implements timer.BackupRestorer
func (s QiniuBackupRestorer) Upload(path string) error {
	name := filepath.Base(path)

	// upload file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	uploadParams := qiniu.UploadParams{
		Name:           filepath.Join("blog", name), // blog/eiblog-xx.tar.gz
		Size:           fi.Size(),
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

// Download implements timer.BackupRestorer
func (s QiniuBackupRestorer) Download() (string, error) {
	// backup file
	params := qiniu.ContentParams{
		Prefix: "blog/",
	}
	raw, err := internal.QiniuClient.Content(params)
	if err != nil {
		return "", err
	}

	path := filepath.Join("/tmp", "eiblog.tar.gz")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	_, _ = f.Write(raw)
	defer f.Close()
	return path, nil
}
