# Logd
高性能日志系统：每日24时自动打包前一天数据，自动删除过期数据（30天），自动邮件告警。

#### Bechmark
```
BenchmarkLogFileChan-8           	  500000	      2265 ns/op
BenchmarkLogFile-8               	  300000	      4393 ns/op
BenchmarkStandardFile-8          	 1000000	      2216 ns/op

BenchmarkLogFileChanMillion-8    	 1000000	      2319 ns/op
BenchmarkLogFileMillion-8        	 1000000	      4397 ns/op
BenchmarkStandardFileMillion-8   	 1000000	      2232 ns/op
```

#### Print Level
```
Ldebug        = 1 << iota //
Linfo                     //
Lwarn                     //
Lerror                    //
Lfatal                    //

logd.Printf()
logd.Print()
logd.Debugf()
logd.Debug()
logd.Infof()
logd.Info()
logd.Warnf()
logd.Warn()
logd.Errorf()
logd.Error()
logd.Fatalf()
logd.Fatal()
```
5种日志等级，12种打印方法，完全理解你记录日志的需求。

```
Ldebug < Linfo < Lwarn < Lerror < Lfatal
```
通过`SetLevel()`方法设置，如：logd.SetLevel(Lwarn)，系统只会打印大于等于Lwarn的日志。

> 注意：logd.Printf() 和 logd.Print() 不论等级为什么都会输出，请悉知。

#### Print Option
系统提供多种格式输出选项：
```
Lasync                    // 异步输出日志
Ldate                     // like 2006/01/02
Ltime                     // like 15:04:05
Lmicroseconds             // like 15:04:05.123123
Llongfile                 // like /a/b/c/d.go:23
Lshortfile                // like d.go:23
LUTC                      // 时间utc输出
Ldaily                    // 按天归档

默认提供：
 LstdFlags // 2006/01/02 15:04:05.123123, /a/b/c/d.go:23
```

#### 使用方法
1、默认方式

默认方式使用以下配置选项：
```
// 2006/01/02 15:04:05.123123, /a/b/c/d.go:23
Lall      = Ldebug | Linfo | Lwarn | Lerror | Lfatal
LstdFlags = Lall | Ldate | Lmicroseconds | Lshortfile
```
这里会打印所有等级日志，打印日志格式为`2006/01/02 15:04:05.123123, /a/b/c/d.go:23`

可以通过`logd.SetLevel(level)`方式调整日志等级。`logd.SetOutput(writer)`输出日志到别的地方。例子：
```
package main

import (
	"fmt"
	"os"

	"github.com/deepzz0/logd"
)

func main() {
	logd.Printf("Printf: foo\n")
	logd.Print("Print: foo")

	logd.Debugf("Debugf: foo\n")
	logd.Debug("Debug: foo")

	logd.Errorf("Errorf: foo\n")
	logd.Error("Error: foo")

	// 改变输出等级
	logd.SetLevel(logd.Lerror)
	// 不论等级如何都会输出
	fmt.Println("----- 改变日志等级为Lerror -----")
	logd.Printf("Printf: foo\n")
	logd.Print("Print: foo")

	logd.Debugf("Debugf: foo\n")
	logd.Debug("Debug: foo")

	logd.Errorf("Errorf: foo\n")
	logd.Error("Error: foo")

	fmt.Println("----- 日志等级为Lerror并写入到test.log中 -----")
	f, err := os.Create("./test.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	logd.SetOutput(f)

	logd.Printf("Printf: foo\n")
	logd.Print("Print: foo")

	logd.Debugf("Debugf: foo\n")
	logd.Debug("Debug: foo")

	logd.Errorf("Errorf: foo\n")
	logd.Error("Error: foo")
}
```
输出如下：
![logd_print](http://7xokm2.com1.z0.glb.clouddn.com/img/logd_print.png)


1、自定义配置
推荐需要将日志保存下来的用户使用。

```
Flags := Lwarn | Lerror | Lfatal | Ldate | Ltime | Lshortfile | Ldaily | Lasync

f, _ := os.OpenFile("testdata/onlyfile.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
log := New(LogOption{
	Out:        f,
	Flag:       Flags,
	LogDir:     "testdata",
	ChannelLen: 1000,
})
```

* 输出日志等级：Lwarn | Lerror | Lfatal
* 输出日志格式：Ldate | Ltime | Lshortfile
* 按日归档：Ldaily
* 异步输出：Lasync

它会将每天的日志，采用 gzip 压缩，并保留 30 天以内的日志。自动删除 30 以前的日志。
