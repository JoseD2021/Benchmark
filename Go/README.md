# Taller Benchmarking – API de órdenes en Go + Postgres

`POST /api/orders` en Go (puerto **5001**) que calcula `totalAmount` e `itemsCount` y guarda en `orders` y `order_items` (Postgres 16, puerto **5433**).

## Requisitos (Windows)
- Docker Desktop (con WSL2) abierto y en *Engine running*
- Git
- Java 17+ y JMeter (zip "Binaries" de jmeter.apache.org), por ejemplo en `C:\apache-jmeter-5.6.3`

No hace falta instalar Go: Docker compila la API.

## 1. Clonar y levantar
```powershell
git clone https://github.com/CaDelTo/taller-benchmarking-go.git
cd taller-benchmarking-go
docker compose up -d --build
docker ps
```

## 2. Probar
En **cmd** (o cualquier terminal, curl viene con Windows 10/11):
```cmd
curl -X POST http://localhost:5001/api/orders -H "Content-Type: application/json" -d "{\"customerId\":\"C123\",\"items\":[{\"productId\":\"P1\",\"quantity\":2,\"price\":10.5},{\"productId\":\"P2\",\"quantity\":1,\"price\":5.0}]}"
```
En **PowerShell**:
```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:5001/api/orders -ContentType "application/json" -Body '{"customerId":"C123","items":[{"productId":"P1","quantity":2,"price":10.5},{"productId":"P2","quantity":1,"price":5.0}]}'
```
Respuesta esperada (201):
```json
{"orderId":"...","customerId":"C123","totalAmount":26,"itemsCount":3,"processedAt":"..."}
```

## 3. JMeter
```powershell
& "C:\apache-jmeter-5.6.3\bin\jmeter.bat" -t orders-test-plan.jmx -Jusers=10
```
- Escenarios: `-Jusers=10`, `50`, `100`; estrés: `20`, `40`, `60`… hasta ver errores o latencia alta.
- Opcionales: `-Jrampup=10` (s), `-Jduration=60` (s).
- Entre corridas: botón *Clear All* (escoba doble).
- Listeners incluidos: Summary Report y Aggregate Report.

## 4. Monitoreo durante la prueba
```powershell
docker stats api-orders db-orders
```
`htop`: desde una terminal WSL (`sudo apt install htop`) o usar el Administrador de tareas.

## 5. Apagar
```powershell
docker compose down      # agrega -v para borrar los datos
```

## Notas
- `itemsCount` = suma de cantidades (3 en el ejemplo).
- Pool de conexiones a Postgres: 50 (`DB_POOL_SIZE` en `docker-compose.yml`).
