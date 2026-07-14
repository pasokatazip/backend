FROM python:3.12-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1 HF_HUB_DISABLE_XET=1

COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "-u", "main.py"]
