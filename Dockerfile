FROM golang:1.25.0-alpine3.21 AS builder

ARG PORT # injected from Railway at build time

ENV APP_HOME /go/src/digimonql

WORKDIR "${APP_HOME}"

COPY . "${APP_HOME}"

RUN go mod download

RUN go mod verify

RUN go build .

FROM scratch
COPY --from=builder /go/src/digimonql/digimonql /
# The file might exist. It might not. If it's there, copy it.
COPY --from=builder /go/src/digimonql/data/digimon.json* /data/

EXPOSE $PORT

CMD ["/digimonql"]
