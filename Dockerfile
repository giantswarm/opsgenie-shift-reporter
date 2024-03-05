FROM alpine:3.11

 RUN apk update && apk --no-cache add ca-certificates && \
  update-ca-certificates

 ADD ./opsgenie-shift-reporter /usr/local/bin/opsgenie-shift-reporter
ENTRYPOINT ["/usr/local/bin/opsgenie-shift-reporter"]
