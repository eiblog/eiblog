package cdn

import (
	"testing"
)

func TestCreateTimestampAntiLeech(t *testing.T) {
	type args struct {
		urlStr            string
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
				urlStr:            "http://www.example.com/testfile.jpg",
				encryptKey:        "abc123",
				durationInSeconds: 3600,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetUrl, err := CreateTimestampAntileechURL(tt.args.urlStr, tt.args.encryptKey, tt.args.durationInSeconds)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTimestampAntiLeech() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(targetUrl)
		})
	}
}
