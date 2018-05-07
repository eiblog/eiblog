.PHONY: test build deploy dist gencert dhparams ssticket makedir clean
# use aliyun dns api to auto renew cert.
# env:
#   export Ali_Key="sdfsdfsdfljlbjkljlkjsdfoiwje"
#   export Ali_Secret="jlsdflanljkljlfdsaklkjflsa"

docker_registry?=registry.cn-hangzhou.aliyuncs.com
acme?=~/.acme.sh
acme.sh?=$(acme)/acme.sh
config?=/data/eiblog/conf


test:

build:
	@echo "go build..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build && \
		docker build -t $(docker_registry)/deepzz/eiblog:latest .

deploy:build
	@docker push $(docker_registry)/deepzz/eiblog:latest

dist:
	@./dist.sh

gencert:makedir
	@if [ ! -n "$(sans)" ]; then \
		printf "Need one argument [sans=params]\n"; \
		printf "example: sans=\"-d domain -d *.domain\"\n"; \
		exit 1; \
	fi; \
	if [ ! -n "$(cn)" ]; then \
		printf "Need one argument [cn=params]\n"; \
		printf "example: cn=domain\n"; \
		exit 1; \
	fi
	@if [ ! -f $(acme.sh) ]; then \
		curl https://get.acme.sh | sh; \
	fi

	@echo "generate rsa cert..."
	@$(acme.sh) --force --issue --dns dns_ali $(sans) \
		--renew-hook "$(acme.sh) --install-cert -d $(cn) \
			--key-file       $(config)/ssl/domain.rsa.key \
			--fullchain-file $(config)/ssl/domain.rsa.pem \
			--reloadcmd      \"service nginx force-reload\""

	@echo "generate ecc cert..."
	@$(acme.sh) --force --issue --dns dns_ali $(sans) -k ec-256 \
		--renew-hook "$(acme.sh) --install-cert -d $(cn) --ecc \
			--key-file       $(config)/ssl/domain.ecc.key \
			--fullchain-file $(config)/ssl/domain.ecc.pem \
			--reloadcmd      \"service nginx force-reload\""

dhparams:
	@openssl dhparam -out $(config)/ssl/dhparams.pem 2048

ssticket:
	@openssl rand 48 > $(config)/ssl/session_ticket.key

makedir:
	@mkdir -p $(config)/ssl

clean:

