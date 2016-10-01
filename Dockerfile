FROM ubuntu:trusty
MAINTAINER deepzz <deepzz.qi@gmail.com>

RUN apt-get update  
RUN apt-get install -y ca-certificates

COPY . /eiblog
EXPOSE 9000
WORKDIR /eiblog
ENTRYPOINT ["./eiblog"]
