
# Instructions for Running the Dockerized Services

This guide provides step-by-step instructions to set up and manage the services defined in the `docker-compose.yaml` file. It covers starting, stopping, and interacting with the services, including the `telemetrygenerator` and `k6` services.

---

## **1. Prerequisites**

- **Docker**: Ensure Docker is installed on your system. [Download Docker](https://www.docker.com/products/docker-desktop).
- **Docker Compose**: Ensure Docker Compose is installed. [Install Docker Compose](https://docs.docker.com/compose/install/).

---

## **2. Setup and Start All Services**

### **Start All Services**

Run the following command to start all services in detached mode:
```bash
docker-compose up -d
```

This will start the following services:
- `db`: PostgreSQL database.
- `telemetryingestion`: UDP listener for telemetry data.
- `turionbackend`: Backend service for telemetry management.
- `turionfrontend`: React frontend for displaying telemetry data.
- `grafana`: Dashboard visualization service.
- `tempo`: Tracing service for telemetry analysis.
- `telemetrygenerator`: Telemetry data generator.
- `k6`: Load testing service (idle by default).

### **Check the Status of Running Services**

To verify that the services are running:
```bash
docker-compose ps
```

You should see the services listed with their `State` as `Up`.

---

## **3. Managing the `telemetrygenerator`**

The `telemetrygenerator` service starts automatically with `docker-compose up`. If you need to stop, restart, or modify its behavior, follow these steps:

### **Stop the Telemetry Generator**

To stop just the `telemetrygenerator`:
```bash
docker-compose stop telemetrygenerator
```

### **Start or Restart the Telemetry Generator**

To restart the telemetry generator:
```bash
docker-compose start telemetrygenerator
```

To stop and then restart it:
```bash
docker-compose restart telemetrygenerator
```

### **Change Packet Delay**

If you want to modify the packet delay:
1. Open the `docker-compose.yml` file.
2. Update the `PACKET_DELAY` environment variable under the `telemetrygenerator` service:
   ```yaml
   environment:
     PACKET_DELAY: "1s"
   ```
3. Apply the changes by restarting the service:
   ```bash
   docker-compose restart telemetrygenerator
   ```

---

## **4. Running `k6` for Load Testing**

The `k6` service is idle by default. You can manually run it to execute load tests.

### **Run the `k6` Service**

To run a load test with the `k6` service:
```bash
docker-compose run --rm k6 k6 run /scripts/k6test.js
```

- **`/scripts/k6test.js`**: Points to the load testing script mounted in the container.

### **Stop the `k6` Service**

Since the `k6` service runs interactively, you can stop it by pressing `Ctrl+C`.

---

## **5. Accessing Services**

### **Frontend Application**
- URL: [http://localhost:3000](http://localhost:3000)

### **Grafana Dashboard**
- URL: [http://localhost:3001](http://localhost:3001)
- Credentials:
    - Username: `admin`
    - Password: `admin`

### **Database**
You can connect to the PostgreSQL database with:
- Host: `localhost`
- Port: `5432`
- Username: `user`
- Password: `password`
- Database: `telemetry`

---

## **6. Backend API Endpoints**

The backend provides the following REST and WebSocket endpoints:

| Endpoint                      | Description                                  | Example URL                                                                                                           | Parameters                                                                                     |
|-------------------------------|----------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------|
| **GET /api/v1/telemetry**     | Retrieve all telemetry data.                | [http://localhost:4000/api/v1/telemetry](http://localhost:4000/api/v1/telemetry)                                      | `start_time` (required, ISO8601), `end_time` (required, ISO8601)                              |
| **GET /api/v1/telemetry/current** | Retrieve the latest telemetry data.      | [http://localhost:4000/api/v1/telemetry/current](http://localhost:4000/api/v1/telemetry/current)                      | No parameters required.                                                                       |
| **GET /api/v1/telemetry/anomalies** | Retrieve telemetry anomalies.          | [http://localhost:4000/api/v1/telemetry/anomalies?start_time=<start>&end_time=<end>](http://localhost:4000/api/v1/telemetry/anomalies) | `start_time` (required, ISO8601), `end_time` (required, ISO8601)                              |
| **GET /api/v1/telemetry/aggregations** | Retrieve aggregated telemetry data. | [http://localhost:4000/api/v1/telemetry/aggregations?start_time=<start>&end_time=<end>&aggregation=<agg>](http://localhost:4000/api/v1/telemetry/aggregations) | `start_time` (required, ISO8601), `end_time` (required, ISO8601), `aggregation` (`min`, `max`, `avg`) |
| **GET /api/v1/telemetry/ws**  | WebSocket endpoint for real-time telemetry. | [ws://localhost:4000/api/v1/telemetry/ws](ws://localhost:4000/api/v1/telemetry/ws)                                    | No parameters required.                                                                       |

---

## **7. Stopping All Services**

To stop all running services:
```bash
docker-compose down
```

This command stops all containers and removes the associated networks, but it will preserve named volumes like `telemetry_db_data` and `grafana-data`.

---

## **8. Cleaning Up**

If you want to remove all containers, networks, and volumes (data will be lost):
```bash
docker-compose down -v
```

---

## **9. Summary of Key Commands**

| Action                           | Command                                                                                 |
|----------------------------------|-----------------------------------------------------------------------------------------|
| Start all services               | `docker-compose up -d`                                                                  |
| Stop all services                | `docker-compose down`                                                                   |
| Restart a specific service       | `docker-compose restart <service_name>`                                                |
| Start the telemetry generator    | `docker-compose start telemetrygenerator`                                               |
| Stop the telemetry generator     | `docker-compose stop telemetrygenerator`                                                |
| Run the k6 load test             | `docker-compose run --rm k6 k6 run /scripts/k6test.js`                                  |
| Check service status             | `docker-compose ps`                                                                     |
| Remove all services and volumes  | `docker-compose down -v`                                                                |

This guide ensures that users can manage all services effectively, while running on-demand tasks like telemetry generation and load testing as needed. 🚀
