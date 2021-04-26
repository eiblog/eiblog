.PHONY: demo build swag

REGISTRY=localhost
IMAGE_TAG=`git describe --tags`

swag:
	@scripts/swag_init.sh

_app:
	@scripts/new_app.sh

# below you should write

# run blog app
blog:
	@scripts/run_app.sh blog

# run backup app
backup:
	@scripts/run_app.sh backup

# build docker
build:
	@scripts/run_build.sh $(REGISTRY) $(IMAGE_TAG)

# protoc
protoc:
	@cd pkg/proto && make protoc
