# Orders API — Java + Spring Boot

Microservicio del taller *Benchmarking de APIs con Postgres usando JMeter*.

- **Stack:** Java 21, Spring Boot 4, Spring JDBC (`JdbcTemplate`), HikariCP, PostgreSQL 16
- **Endpoint:** `POST http://localhost:8080/api/orders`

## Levantar el entorno

```bash
docker compose up -d --build
docker ps
```

Las tablas `orders` y `order_items` se crean con `db/init.sql` (y también al arrancar la API vía `schema.sql`).

## Probar

```bash
curl -s -X POST http://localhost:8080/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"customerId":"C123","items":[{"productId":"P1","quantity":2,"price":10.5},{"productId":"P2","quantity":1,"price":5.0}]}'
```

Respuesta (`201 Created`):

```json
{
  "orderId": "3f6c1e0a-...",
  "customerId": "C123",
  "totalAmount": 26.00,
  "itemsCount": 3,
  "processedAt": "2026-11-20T10:15:30.123"
}
```

- `totalAmount` = Σ (quantity × price)
- `itemsCount` = Σ quantity (unidades totales)
- La orden y sus items se insertan en una sola transacción (los items con *batch insert*).

Verificar en la BD:

```bash
docker exec -it db-orders psql -U postgres -d orders_db -c "SELECT count(*) FROM orders;"
```

## Pruebas de carga con JMeter

El plan `jmeter/orders-test-plan.jmx` incluye: Thread Group, HTTP Request Defaults (localhost:8080, `/api/orders`), HTTP Header Manager (`Content-Type: application/json`), HTTP Request POST con el JSON, aserción de status 201, Summary Report y Aggregate Report.

Los usuarios, ramp-up y duración son propiedades (`-Jusers`, `-Jrampup`, `-Jduration`), por defecto 10 usuarios / 10 s / 60 s.

**Con interfaz gráfica** (para los pantallazos):

```bash
jmeter -t jmeter/orders-test-plan.jmx -Jusers=50
```

**En CLI** (genera reporte HTML con promedio, P90/P95/P99, throughput y % error):

```bash
./jmeter/run-scenarios.sh                       # 10, 50 y 100 usuarios
./jmeter/run-scenarios.sh 120 140 160 180 200   # estrés, de 20 en 20
```

Mientras corre cada prueba, en otras terminales: `htop` y `docker stats`.

> Entre escenarios se puede limpiar la BD: `docker exec db-orders psql -U postgres -d orders_db -c "TRUNCATE orders, order_items;"`

## Resultados (Spring Boot)

**Entorno:** 16 núcleos, 31 GiB RAM, Arch Linux. API y Postgres en Docker; JMeter 5.6 en CLI en la misma máquina.
Cada escenario: ramp-up de 10 s, 60 s de duración, loop infinito (sin think time). Tablas truncadas antes de cada escenario.
Pool Hikari: 20 conexiones. Tomcat: 200 hilos. Datos crudos en `jmeter/results/summary.csv` y reportes HTML en `jmeter/results/<N>u/index.html`.

### Desempeño

| Escenario | Requests | Avg (ms) | P90 (ms) | P95 (ms) | P99 (ms) | Throughput (req/s) | % Error | Observaciones |
|---|---|---|---|---|---|---|---|---|
| Baja (10) | 202,152 | 2.71 | 3 | 3 | 3 | 3,372 | 0% | Muy estable; el límite es la cantidad de usuarios, no el servidor |
| Media (50) | 417,426 | 6.56 | 8 | 9 | 9 | **6,962** | 0% | Mejor throughput de todas las pruebas |
| Alta (100) | 226,962 | 24.22 | 26 | 26 | 45 | 3,785 | 0% | El throughput cae ~45% aunque la CPU baja: contención por el pool de 20 conexiones |

### Estrés (aumentando de 20 en 20)

| Usuarios | Avg (ms) | P95 (ms) | P99 (ms) | Max (ms) | Throughput (req/s) | % Error |
|---|---|---|---|---|---|---|
| 120 | 29.13 | 33 | 35 | 168 | 3,779 | 0% |
| 140 | 35.32 | 42 | 78 | 187 | 3,635 | 0% |
| 160 | 38.71 | 43 | 54 | 193 | 3,787 | 0% |
| 180 | 43.35 | 60 | 95 | 204 | 3,803 | 0% |
| 200 | 47.36 | 52 | 64 | 211 | 3,872 | 0% |
| 300 | 72.88 | 173 | 221 | 605 | 3,777 | 0% |
| 500 | 118.09 | 204 | 258 | 582 | 3,881 | 0% |

**Punto de estrés:** a partir de **~100 usuarios** el sistema se satura. El throughput se estanca en ~3,800 req/s y la latencia crece de forma casi lineal con los usuarios: los requests extra solo hacen cola. Por encima de 200 usuarios (más que los hilos de Tomcat) el P95 se dispara (173–204 ms). No hubo errores ni timeouts en ningún escenario, ni siquiera con 500 usuarios.

### Consumo del sistema (`docker stats`, muestreado cada ~2 s)

CPU en % de un núcleo (100% = 1 núcleo de 16).

| Escenario | CPU API prom / máx (%) | RAM API máx (MiB) | CPU DB prom / máx (%) | RAM DB máx (MiB) | Observaciones |
|---|---|---|---|---|---|
| Baja (10) | 48 / 218 | 378 | 40 / 66 | 132 | Pico de CPU por el JIT al inicio, luego estable |
| Media (50) | 89 / 120 | 375 | 88 / 112 | 207 | Máximo uso de CPU = máximo throughput |
| Alta (100) | 52 / 62 | 389 | 50 / 59 | 205 | Menos CPU con más carga: los hilos esperan conexión, no trabajan |
| Estrés (200) | 61 / 79 | 400 | 53 / 74 | 221 | Igual que con 100: saturado por contención, no por recursos |
| Estrés (500) | 59 / 74 | 419 | 54 / 74 | 215 | RAM de la API casi constante (~400 MiB) |

### Observaciones

- **Sin errores en ninguna carga.** La JVM, el pool de conexiones y la cola de Tomcat absorben la carga encolando en vez de fallar.
- **El cuello de botella no es CPU ni RAM** (la máquina tiene 16 núcleos y la API nunca pasó de ~2.2 núcleos). Es la **contención por el pool de 20 conexiones**: con 50 usuarios hay ~2.5 hilos por conexión y todo fluye; con 100 o más, decenas de hilos compiten por cada conexión y el costo de coordinarlos (*handoff*) reduce el throughput. Subir `DB_POOL_SIZE` o usar virtual threads probablemente movería este límite.
- **Memoria muy estable** (~375–420 MiB): el GC de la JVM no causó pausas visibles en la latencia.
- **Calentamiento del JIT:** se hizo una corrida previa de calentamiento; en la primera prueba de 10 usuarios todavía hay un pico de CPU de 218% mientras se compila el código caliente.
- JMeter corre en la misma máquina, así que también consume CPU y compite con la API en las cargas altas.
