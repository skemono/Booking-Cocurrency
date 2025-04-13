# 🎟️ Booking Concurrency — Simulación de Reservas Concurrentes

¡Bienvenido a **Booking-Concurrency**!  
Este proyecto simula múltiples usuarios intentando **reservar asientos para eventos** al mismo tiempo — perfecto para **practicar operaciones ACID en PostgreSQL** usando **Go (Golang)** y **goroutines** ⚙️🐿️

---

## 📚 ¿Qué hace este proyecto?

🧠 El objetivo es simular un escenario real donde **muchas personas intentan reservar el mismo asiento** para un evento, al mismo tiempo.  
Esto te permite poner a prueba:

- ✅ Transacciones atómicas
- 🔒 Bloqueos de fila (row-level locking)
- 🧱 Integridad de datos
- ⚡ Concurrencia realista con goroutines

---

## 🛠️ Tecnologías

- 🐘 **PostgreSQL 15** (Dockerizado)
- 🧠 **Go (Golang)**
- 🗂️ **Goose** para migraciones SQL
- 🧵 **Goroutines** para concurrencia

---

## 🚀 Cómo levantar la base de datos (PostgreSQL con Docker)

Asegúrate de tener [Docker](https://www.docker.com/) instalado.

1. En la raíz del proyecto, corre:

   ```bash
   docker-compose up -d
   ```

2. PostgreSQL estará corriendo en el puerto `5432` (Si tienes una instancia de PostgreSQL inicializada la mayoría de veces es necesario apagarla en Servicios para que los servidores no colisionen).  
   Credenciales:

   ```
   Host: localhost
   User: admin
   Password: reservasDB123!
   DB: reservas_db
   ```

---

## 🧬 Estructura del Proyecto

```
📁 Booking-Concurrency/
│
├── 📁 db/
│   ├── 001_ddl.sql         # Esquema de tablas
│   └── 002_dml.sql         # Datos iniciales de eventos y asientos
│
├── docker-compose.yml      # PostgreSQL container
├── go.mod / go.sum         # Dependencias de Go
└── main.go                 # Simulación concurrente de reservas
```
