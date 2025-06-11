FROM python:3.12-alpine

WORKDIR /app

COPY server.py .

RUN pip install --no-cache-dir aiohttp==3.*

EXPOSE 8080

ENTRYPOINT ["python", "server.py"]