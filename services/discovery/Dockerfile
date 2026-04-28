FROM python:3.12-alpine
WORKDIR /app
ENV PYTHONUNBUFFERED=1
COPY dns.py /app/dns.py
COPY cluster_config.yaml /app/cluster_config.yaml
EXPOSE 6699/udp
CMD ["python3", "/app/dns.py"]
