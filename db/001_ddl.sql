-- +goose Up
-- SQL in this section is executed when the migration is applied
CREATE TABLE IF NOT EXISTS Eventos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    fecha TIMESTAMP NOT NULL,
    lugar VARCHAR(100) NOT NULL
);

-- Tabla de Asientos
CREATE TABLE IF NOT EXISTS Asientos (
    id SERIAL PRIMARY KEY,
    evento_id INTEGER NOT NULL REFERENCES Eventos(id) ON DELETE CASCADE,
    numero_asiento INTEGER NOT NULL,
    estado VARCHAR(20) NOT NULL CHECK (estado IN ('disponible', 'reservado')),
    UNIQUE (evento_id, numero_asiento)
);

-- Tabla de Reservas
CREATE TABLE IF NOT EXISTS Reservas (
    id SERIAL PRIMARY KEY,
    usuario_id VARCHAR(50) NOT NULL,
    asiento_id INTEGER NOT NULL REFERENCES Asientos(id) ON DELETE CASCADE,
    fecha_reserva TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estado_reserva VARCHAR(20) NOT NULL CHECK (estado_reserva IN ('exitosa', 'fallida'))
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back

DROP TABLE IF EXISTS Reservas;
DROP TABLE IF EXISTS Asientos;
DROP TABLE IF EXISTS Eventos;