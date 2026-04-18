FROM python:3.12-alpine
WORKDIR /app
ENV PYTHONUNBUFFERED=1
COPY dns.py /app/dns.py
EXPOSE 6699/udp
CMD ["python3", "/app/dns.py"]
