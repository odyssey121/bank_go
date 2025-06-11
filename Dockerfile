FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz | tar xvz
RUN ls -la
RUN go build -o main main.go


FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
RUN mv /app/migrate /usr/bin/migrate -v
RUN chmod 777 /usr/bin/migrate
COPY config_docker.yaml .
COPY db/migration ./migration
COPY start.sh ./start.sh
RUN chmod 777 /app/start.sh
COPY wait-for.sh ./wait-for.sh
RUN chmod 777 /app/wait-for.sh

RUN ls -la


EXPOSE 8888

CMD [ "/app/main" ]
ENTRYPOINT [ "/app/wait-for.sh", "postgres_container:5432", "--", "/app/start.sh" ]