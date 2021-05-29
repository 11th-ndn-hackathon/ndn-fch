FROM golang:1.16.4-buster AS build
WORKDIR /go/src/app
COPY . .
RUN env CGO_ENABLED=0 go build .

FROM scratch
COPY --from=build /go/src/app/ndn-fch /ndn-fch
ENTRYPOINT ["/ndn-fch"]
