FROM alpine:latest
LABEL maintainer="deepzz.qi@gmail.com"

RUN apk add --update --no-cache tzdata ca-certificates mongodb-tools
COPY README.md /app/README.md
COPY CHANGELOG.md /app/CHANGELOG.md
COPY LICENSE /app/LICENSE

COPY bin/backend /app/backend
COPY conf /app/conf

EXPOSE 9001

WORKDIR /app
CMD ["./backend"]
