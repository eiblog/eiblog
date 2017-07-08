package logd

import (
	"log"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	Printf("Printf: foo\n")
	Print("Print: foo")

	SetLevel(Ldebug)

	Debugf("Debugf: foo\n")
	Debug("Debug: foo")

	Infof("Infof: foo\n")
	Info("Info: foo")

	Errorf("Errorf: foo\n")
	Error("Error: foo")

	SetLevel(Lerror)

	Debugf("Debugf: foo\n")
	Debug("Debug: foo")

	Infof("Infof: foo\n")
	Info("Info: foo")

	Errorf("Errorf: foo\n")
	Error("Error: foo")
}

func BenchmarkLogFileChan(b *testing.B) {
	log := New(LogOption{
		Flag:       LAsync | Ldate | Ltime | Lshortfile,
		LogDir:     "testdata",
		ChannelLen: 1000,
	})

	for i := 0; i < b.N; i++ {
		log.Print("testing this is a testing about benchmark")
	}
	log.WaitFlush()
}

func BenchmarkLogFile(b *testing.B) {
	f, _ := os.OpenFile("testdata/onlyfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	log := New(LogOption{
		Out:        f,
		Flag:       Ldate | Ltime | Lshortfile,
		LogDir:     "testdata",
		ChannelLen: 1000,
	})

	for i := 0; i < b.N; i++ {
		log.Print("testing this is a testing about benchmark")
	}
	log.WaitFlush()
}

func BenchmarkStandardFile(b *testing.B) {
	f, _ := os.OpenFile("testdata/logfile.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	log := log.New(f, "", log.LstdFlags)
	for i := 0; i < b.N; i++ {
		log.Print("testing this is a testing about benchmark")
	}
}

func BenchmarkLogFileChanMillion(b *testing.B) {
	log := New(LogOption{
		Flag:       LAsync | Ldate | Ltime | Lshortfile,
		LogDir:     "testdata",
		ChannelLen: 1000,
	})
	b.N = 1000000
	for i := 0; i < b.N; i++ {
		log.Print("testing this is a testing about benchmark")
	}
	log.WaitFlush()
}

func BenchmarkLogFileMillion(b *testing.B) {
	f, _ := os.OpenFile("testdata/onlyfilemillion.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	log := New(LogOption{
		Out:        f,
		Flag:       Ldate | Ltime | Lshortfile,
		LogDir:     "testdata",
		ChannelLen: 1000,
	})
	b.N = 1000000
	for i := 0; i < b.N; i++ {
		log.Print("testing this is a testing about benchmark")
	}
	log.WaitFlush()
}

func BenchmarkStandardFileMillion(b *testing.B) {
	f, _ := os.OpenFile("testdata/logfilemillion.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	log := log.New(f, "", log.LstdFlags)
	b.N = 1000000
	for i := 0; i < b.N; i++ {
		log.Print("testing this is a testing about benchmark")
	}
}
