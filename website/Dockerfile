# Public Domain (-) 2020-present, The Core Authors.
# See the Core UNLICENSE file for details.

FROM golang:1.13-buster as builder
ADD . /src/
ENV CGO_ENABLED=0
WORKDIR /src
RUN ./environ/selfsigned.sh
RUN go build -o /app ./app

FROM scratch
COPY --from=builder /app /
COPY --from=builder /src/selfsigned/tls.cert /selfsigned/tls.cert
COPY --from=builder /src/selfsigned/tls.key /selfsigned/tls.key
ENV PRODUCTION=1
ENTRYPOINT ["/app"]
