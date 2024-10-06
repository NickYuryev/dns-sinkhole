FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /sinkhole .

FROM alpine:3.20
COPY --from=build /sinkhole /sinkhole
COPY blocklist.txt /blocklist.txt
EXPOSE 8053/udp 9090/tcp
ENTRYPOINT ["/sinkhole"]
CMD ["--blocklist=/blocklist.txt"]
