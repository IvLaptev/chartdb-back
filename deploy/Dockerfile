FROM bufbuild/buf as generator

WORKDIR /gen

COPY ./app .
RUN buf dep update

RUN buf generate

FROM golang:alpine as builder

WORKDIR /build

COPY --from=generator /gen/api ./api

ADD ./app/go.mod .
COPY ./app .

RUN go build -o chartdb_back ./cmd/chartdb_back

FROM alpine

WORKDIR /app

RUN apk --no-cache add curl

COPY --from=builder /build/chartdb_back /app/chartdb_back
COPY ./app/pkg/emailsender/templates/* /emailsender/templates/

CMD [ "./chartdb_back", "-c", "config.yaml" ]
