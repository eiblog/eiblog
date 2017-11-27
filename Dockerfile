FROM alpine
MAINTAINER deepzz <deepzz.qi@gmail.com>

RUN apk add --update --no-cache ca-certificates
ADD static/tzdata/Shanghai /etc/localtime

COPY . /eiblog
EXPOSE 9000
WORKDIR /eiblog
CMD ["sh","-c","/eiblog/eiblog"]
