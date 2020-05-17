FROM golang:alpine AS builder
RUN apk update && apk add git
RUN go get github.com/rakyll/statik
RUN git clone https://github.com/mhlo/tesk.git
WORKDIR $GOPATH/src/github.com/mhlo/tesk
RUN adduser -S -D -H -h /go/src/tesk appuser
# COPY main.go /go/src/tesk
# RUN /go/bin/statik -f -src /tmp/files-statik/; go generate;


FROM alpine
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/statik /go/bin/statik
USER appuser
# CMD /go/bin/statik -src /files-statik/ -dest /dest/
ENTRYPOINT ["/go/bin/statik", "-f", "-src", "/files-statik/", "-dest", "/dest/"]
