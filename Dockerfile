FROM ubuntu:trusty
MAINTAINER deepzz <deepzz.qi@gmail.com>

RUN apt-get update  
RUN apt-get install -y ca-certificates

ADD conf /eiblog/conf
ADD static /eiblog/static
ADD views /eiblog/views
ADD eiblog /eiblog/eiblog

EXPOSE 9000

WORKDIR /eiblog
ENTRYPOINT ["./eiblog"]