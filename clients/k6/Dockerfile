FROM grafana/k6:latest
USER root
RUN apk add --no-cache python3
WORKDIR /app
COPY clients/discovery-proxy/service_discovery_client.py /app/service_discovery_client.py
COPY clients/k6 /app/k6
COPY clients/k6/entrypoint.sh /app/entrypoint.sh
COPY cluster_config.yaml /app/cluster_config.yaml
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
