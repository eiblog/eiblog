FROM alpine
MAINTAINER deepzz <deepzz.qi@gmail.com>

RUN apk update
RUN apk add ca-certificates
RUN echo "Asia/Shanghai" > /etc/timezone

COPY . /eiblog
EXPOSE 9000
WORKDIR /eiblog
ENTRYPOINT ["./eiblog"]
