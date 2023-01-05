FROM alpine:latest
LABEL maintainer="deepzz.qi@gmail.com"

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories \
    && apk add --update --no-cache tzdata ca-certificates mongodb-tools libc6-compat
COPY README.md /app/README.md
COPY CHANGELOG.md /app/CHANGELOG.md
COPY LICENSE /app/LICENSE

COPY bin/backend /app/backend
COPY conf /app/conf

EXPOSE 9001

WORKDIR /app
CMD ["./backend"]
