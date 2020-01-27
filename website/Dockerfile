# Public Domain (-) 2020-present, The Core Authors.
# See the Core UNLICENSE file for details.

FROM golang:1.13-buster as builder
ADD . /src/
ENV CGO_ENABLED=0
WORKDIR /src
RUN go build -o /app ./app

FROM scratch
COPY --from=builder /app /
ENV PORT=8080
ENTRYPOINT ["/app"]
