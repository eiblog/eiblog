FROM alpine:latest
LABEL maintainer="deepzz.qi@gmail.com"

COPY README.md /app/README.md
COPY CHANGELOG.md /app/CHANGELOG.md
COPY LICENSE /app/LICENSE

COPY bin/backend /app/backend
COPY conf /app/conf

EXPOSE 9000

WORKDIR /app
CMD ["backend"]
