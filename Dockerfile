FROM alpine:3.10
RUN apk --no-cache add ca-certificates tzdata
#COPY script/ca-certificates.crt /etc/ssl/certs/
COPY dist/traefik /usr/local/bin/
RUN chmod +x /usr/local/bin/traefik
COPY entrypoint.sh /
EXPOSE 80
VOLUME ["/tmp"]
ENTRYPOINT ["/entrypoint.sh"]
CMD ["traefik"]

# Metadata
LABEL org.opencontainers.image.vendor="Containous" \
	org.opencontainers.image.url="https://traefik.io" \
	org.opencontainers.image.title="Traefik" \
	org.opencontainers.image.description="A modern reverse-proxy" \
	org.opencontainers.image.version="v1.7" \
	org.opencontainers.image.documentation="https://docs.traefik.io"