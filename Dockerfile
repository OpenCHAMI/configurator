FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini bash
RUN mkdir -p /configurator
RUN mkdir -p /configurator/lib

# nobody 65534:65534
USER 65534:65534

# copy the binary and all of the default plugins
COPY configurator /configurator/configurator
COPY lib/* /configurator/lib/*

CMD ["/configurator"]

ENTRYPOINT [ "/sbin/tini", "--" ]