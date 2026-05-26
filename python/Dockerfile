FROM python:3.12-slim

WORKDIR /app

# Python依存関係をコピー
COPY requirements.txt ./

# インストール
RUN pip install --no-cache-dir -r requirements.txt

# Pythonアプリコピー
COPY main.py ./

EXPOSE 8000

# FastAPIを起動
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]