import os
import time
import uuid
from datetime import datetime, timezone

import psycopg2
from flask import Flask, jsonify, request
from psycopg2.pool import SimpleConnectionPool

app = Flask(__name__)

DB = dict(
    host=os.getenv("DB_HOST", "localhost"),
    port=os.getenv("DB_PORT", "5432"),
    user=os.getenv("DB_USER", "postgres"),
    password=os.getenv("DB_PASSWORD", "postgres"),
    dbname=os.getenv("DB_NAME", "orders_db"),
)

SCHEMA = """
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) UNIQUE NOT NULL,
    customer_id VARCHAR(50) NOT NULL,
    total_amount NUMERIC(10,2) NOT NULL,
    items_count INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    price NUMERIC(10,2) NOT NULL
);
"""


def connect_pool():
    # La BD puede tardar unos segundos en aceptar conexiones: reintenta
    for _ in range(30):
        try:
            return SimpleConnectionPool(1, int(os.getenv("DB_POOL_MAX", "1")), **DB)
        except psycopg2.OperationalError:
            time.sleep(1)
    raise RuntimeError("no se pudo conectar a la base de datos")


# Un pool por worker (cada worker de gunicorn atiende una peticion a la vez, sin hilos)
pool = connect_pool()


def create_tables():
    conn = pool.getconn()
    try:
        with conn, conn.cursor() as cur:
            # El lock evita que los workers choquen creando las tablas a la vez
            cur.execute("SELECT pg_advisory_xact_lock(1)")
            cur.execute(SCHEMA)
    finally:
        pool.putconn(conn)


create_tables()


@app.post("/api/orders")
def create_order():
    data = request.get_json(silent=True)
    if not data or not data.get("customerId") or not data.get("items"):
        return jsonify(error="customerId e items son obligatorios"), 400

    customer_id = data["customerId"]
    items = data["items"]

    # 1. Calcular totales
    try:
        total_amount = round(sum(i["quantity"] * i["price"] for i in items), 2)
    except (KeyError, TypeError):
        return jsonify(error="cada item necesita productId, quantity y price"), 400
    items_count = len(items)

    order_id = "ORD-" + uuid.uuid4().hex[:16]
    now = datetime.now(timezone.utc)

    # 2. Insertar en orders y order_items (misma transaccion)
    conn = pool.getconn()
    try:
        with conn, conn.cursor() as cur:
            cur.execute(
                "INSERT INTO orders (order_id, customer_id, total_amount, items_count, created_at) "
                "VALUES (%s, %s, %s, %s, %s)",
                (order_id, customer_id, total_amount, items_count, now),
            )
            cur.executemany(
                "INSERT INTO order_items (order_id, product_id, quantity, price) "
                "VALUES (%s, %s, %s, %s)",
                [(order_id, i["productId"], i["quantity"], i["price"]) for i in items],
            )
    except Exception as e:
        app.logger.error("error guardando la orden: %s", e)
        return jsonify(error="error interno"), 500
    finally:
        pool.putconn(conn)

    # 3. Respuesta
    return jsonify(
        orderId=order_id,
        customerId=customer_id,
        totalAmount=total_amount,
        itemsCount=items_count,
        processedAt=now.isoformat(),
    ), 201
