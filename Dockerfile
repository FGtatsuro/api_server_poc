FROM golang:1.16 as builder
WORKDIR /app

COPY go.mod ./
COPY main.go ./

# To avoid dependence on libc
ARG CGO_ENABLED=0
RUN go mod download
RUN go build -v -o server


FROM scratch
COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
