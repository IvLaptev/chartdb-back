FROM golang:alpine as builder

WORKDIR /build

ADD ./app/go.mod .
COPY ./app .

RUN go build -o chartdb_back ./cmd/chartdb_back

FROM alpine

WORKDIR /app

COPY --from=builder /build/chartdb_back /app/chartdb_back

CMD [ "./chartdb_back", "-c", "config.yaml" ]
