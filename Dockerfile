FROM golang:1.15 as build
WORKDIR $GOPATH/avalanche-query
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o=/bin/query ./cmd

FROM scratch
COPY --from=build /bin/query /bin/query
EXPOSE 9001
ENTRYPOINT ["/bin/query"]
