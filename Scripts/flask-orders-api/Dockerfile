FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY app.py .
EXPOSE 5001
# 4 workers (procesos), sin hilos: cada worker atiende una peticion a la vez
CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:5001", "app:app"]
