# API Orders - Python + Flask

Endpoint: `POST http://localhost:5001/api/orders`

## Levantar

```bash
docker compose up -d --build
docker ps
```

Las tablas `orders` y `order_items` se crean solas la primera vez (desde `init.sql`).

## Probar

```bash
curl -X POST http://localhost:5001/api/orders \
  -H "Content-Type: application/json" \
  -d '{"customerId":"C123","items":[{"productId":"P1","quantity":2,"price":10.5},{"productId":"P2","quantity":1,"price":5.0}]}'
```

Respuesta esperada (201):

```json
{"orderId":"ORD-...","customerId":"C123","totalAmount":26.0,"itemsCount":2,"processedAt":"2026-..."}
```

## Notas

- `itemsCount` = numero de lineas en `items` (2 en el ejemplo). Si el profesor lo quiere como suma de cantidades, cambiar `len(items)` en `app.py`.
- Servidor: gunicorn con 4 workers (procesos) y sin hilos (ver `Dockerfile`). Cada worker usa 1 conexion a Postgres.
- Las tablas se crean al hacer `docker compose up` (desde `init.sql`) y, ademas, la API las crea al arrancar si no existen (por si el volumen ya existia).
- Para reiniciar la BD desde cero: `docker compose down -v`.
- Monitoreo durante la prueba: `docker stats` y `htop`.
