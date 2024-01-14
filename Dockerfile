FROM golang:1.21-alpine3.19 AS build
WORKDIR /app
COPY . .
RUN env CGO_ENABLED=0 GOBIN=/build go install ./cmd/...

FROM scratch
COPY --from=build /build/* /
WORKDIR /runtime
ENTRYPOINT ["/ndn-fch-api"]
