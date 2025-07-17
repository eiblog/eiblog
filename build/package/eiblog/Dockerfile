FROM golang:1.20 AS builder

WORKDIR /eiblog
COPY . .
RUN ./scripts/run_build.sh eiblog


FROM alpine:latest
LABEL maintainer="deepzz.qi@gmail.com"

RUN apk add --update --no-cache tzdata
COPY README.md /app/README.md
COPY CHANGELOG.md /app/CHANGELOG.md
COPY LICENSE /app/LICENSE

COPY --from=builder /eiblog/cmd/eiblog/backend /app/backend
COPY cmd/eiblog/etc /app/etc

EXPOSE 9000

WORKDIR /app
CMD ["./backend"]
