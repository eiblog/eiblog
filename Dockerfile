FROM ubuntu:trusty
MAINTAINER deepzz <deepzz.qi@gmail.com>

ADD conf /eiblog/conf
ADD static /eiblog/static
ADD views /eiblog/views
ADD eiblog /eiblog/eiblog

EXPOSE 80
EXPOSE 443

WORKDIR /eiblog
ENTRYPOINT ["./eiblog"]