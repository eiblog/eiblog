FROM alpine:latest
LABEL maintainer="deepzz.qi@gmail.com"

RUN apk add --update --no-cache tzdata
COPY README.md /app/README.md
COPY CHANGELOG.md /app/CHANGELOG.md
COPY LICENSE /app/LICENSE

COPY bin/backend /app/backend
COPY conf /app/conf
COPY website /app/website
COPY assets /app/assets

EXPOSE 9000

WORKDIR /app
CMD ["./backend"]
