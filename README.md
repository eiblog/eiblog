# APP Demo

Since there are many Web apps developed, a common development template is put together here.

### New app
1、You need copy `appdemo` to your `GOPATH` and rename:
```
$ git clone git@github.com:deepzz0/appdemo.git <app name>
```

3、Enter your app, run:
```
$ cd <app name>
$ make _app
```

3、Push the code to new repo:
```
$ git add .
$ git commit -m "init repo"
$ git remote add origin <your repo>
$ git push -u origin master
```

4、`make demo` you can start your web app.

### Development

**Step1**

Understand the directory.

```
.
├── build             # Packaging and CI.
├── cmd               # Main applications for this app.
├── conf              # Static configuration file.
├── docs              # Design and user documents.
├── pkg               # Library code that's ok to use by external applications.
├── scripts           # Scripts to perform various build, install, analysis, etc operations.
├── website           # APP's website data.
├── CHANGELOG.md      # Record version change.
├── LICENSE           # Open source license
├── Makefile          # Makefile: call scripts
├── README.md         # Read me docs.
└── go.mod            # Go mod file.
```



**Step2**

Code in pkg and cmd or website.



