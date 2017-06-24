# use aliyun dns api to auto renew cert.
# env:
#   export Ali_Key="sdfsdfsdfljlbjkljlkjsdfoiwje"
#   export Ali_Secret="jlsdflanljkljlfdsaklkjflsa"

docker_registry?=registry.cn-hangzhou.aliyuncs.com
acme?=~/.acme.sh
acme.sh?=$(acme)/acme.sh
config?=tmp/conf


test:

build:
	@echo "go build..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build && \
	  docker build -t $(docker_registry)/deepzz/eiblog .

deploy:build
	@docker push $(docker_registry)/deepzz/eiblog

dist:
	@./dist.sh

gencert:
	@echo $(Ali_Key) $(Ali_Secret)
	@if [ ! -n "$(sans)" ]; then \
      printf "Need one argument [sans=params]\n"; \
	  printf "example: sans=\"-d domain -d domain\"\n"; \
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
	@$(acme.sh) --force --issue --dns dns_ali \
	  $(sans) --log
	@$(acme.sh) --install-cert -d $(cn) \
      --key-file       $(config)/ssl/domain.rsa.key \
      --fullchain-file $(config)/ssl/domain.rsa.pem \
      --reloadcmd      "service nginx force-reload"
	@ct-submit ctlog.api.venafi.com < $(config)/ssl/domain.rsa.pem > $(config)/scts/rsa/venafi.sct && \
	  ct-submit ctlog-gen2.api.venafi.com < $(config)/ssl/domain.rsa.pem > $(config)/scts/rsa/venafi2.sct

	@echo "generate ecc cert..."
	@$(acme.sh) --force --issue --dns dns_ali \
	  $(sans) -k ec-256 --log
	@$(acme.sh) --install-cert -d $(cn) --ecc \
      --key-file       $(config)/ssl/domain.ecc.key \
      --fullchain-file $(config)/ssl/domain.ecc.pem \
      --reloadcmd      "service nginx force-reload"
	@ct-submit ctlog.api.venafi.com < $(config)/ssl/domain.ecc.pem > $(config)/scts/ecc/venafi.sct && \
	  ct-submit ctlog-gen2.api.venafi.com < $(config)/ssl/domain.ecc.pem > $(config)/scts/ecc/venafi2.sct

# fullchained:
# 	@if [ ! -n "$(cn)" ]; then \
#       printf "Use acme.sh generated certs, Need one argument [cn=params]\n"; \
# 	  printf "example: cn=domain\n"; \
#       exit 1; \
#     fi
# 	@cp $(acme)/$(cn)/ca.cer $(config)/ssl/full_chained.pem && \
# 	   echo $(X3) >> $(config)/ssl/full_chained.pem

dhparams:
	@openssl dhparam -out $(config)/ssl/dhparams.pem 2048

clean:

