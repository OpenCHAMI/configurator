FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini bash

# nobody 65534:65534
USER 65534:65534

# copy the binary and all of the default plugins
COPY configurator /configurator
COPY lib/* /lib/*

CMD ["/configurator"]

ENTRYPOINT [ "/sbin/tini", "--" ]