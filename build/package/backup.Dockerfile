FROM golang:1.20 AS builder

WORKDIR /eiblog
COPY . .
RUN ./scripts/run_build.sh backup


FROM alpine:latest
LABEL maintainer="deepzz.qi@gmail.com"

RUN apk add --update --no-cache tzdata ca-certificates \
  mongodb-tools libc6-compat gcompat
COPY README.md /app/README.md
COPY CHANGELOG.md /app/CHANGELOG.md
COPY LICENSE /app/LICENSE

COPY --from=builder /eiblog/bin/backend /app/backend
COPY conf /app/conf

EXPOSE 9001

WORKDIR /app
CMD ["./backend"]
