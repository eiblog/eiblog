package cdn

import (
	"net/url"
	"testing"
)

func TestCreateTimestampAntiLeech(t *testing.T) {
	type args struct {
		host              string
		fileName          string
		queryStr          url.Values
		encryptKey        string
		durationInSeconds int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "antileech_1",
			args: args{
				host:     "http://www.abc.com",
				fileName: "abc.jpg",
				queryStr: url.Values{
					"x": {"9"},
				},
				encryptKey:        "abc",
				durationInSeconds: 20,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateTimestampAntileechURL(tt.args.host, tt.args.fileName, tt.args.queryStr, tt.args.encryptKey, tt.args.durationInSeconds)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTimestampAntiLeech() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
