.PHONY: demo build swag

tag=`git describe --abbrev=0 --tags`

swag:
	@scripts/swag_init.sh

_app:
	@scripts/new_app.sh

# below you should write

# run eiblog app
eiblog:
	@scripts/run_app.sh eiblog

# run backup app
backup:
	@scripts/run_app.sh backup

# dist tar
dist:
	@scripts/dist_tar.sh $(tag)

# protoc
protoc:
	@cd pkg/proto && make protoc
