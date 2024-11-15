FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini bash
RUN mkdir -p /configurator

# nobody 65534:65534
USER 65534:65534

# copy the binary and all of the default plugins
COPY configurator /configurator/configurator

CMD ["/configurator/configurator"]

ENTRYPOINT [ "/sbin/tini", "--" ]