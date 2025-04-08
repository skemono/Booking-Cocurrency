-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    -- Verificar si la tabla eventos está vacía
    IF NOT EXISTS (SELECT 1 FROM eventos LIMIT 1) THEN
        -- Datos iniciales para Eventos
        INSERT INTO eventos (nombre, fecha, lugar) VALUES
            ('Concierto de Orquesta Sinfónica', '2025-05-15 20:00:00', 'Teatro Nacional'),
            ('Conferencia de Inteligencia Artificial', '2025-06-10 09:00:00', 'Centro de Convenciones'),
            ('Romeo y Julieta', '2025-04-25 19:30:00', 'Teatro Municipal');

        -- Datos iniciales para Asientos
        -- Concierto (50 asientos)
        INSERT INTO asientos (evento_id, numero_asiento, estado)
        SELECT 1, generate_series(1, 50), 'disponible';

        -- Conferencia (30 asientos)
        INSERT INTO asientos (evento_id, numero_asiento, estado)
        SELECT 2, generate_series(1, 30), 'disponible';

        -- Teatro (20 asientos)
        INSERT INTO asientos (evento_id, numero_asiento, estado)
        SELECT 3, generate_series(1, 20), 'disponible';
        
        RAISE NOTICE 'Datos iniciales insertados exitosamente';
    ELSE
        RAISE NOTICE 'Las tablas ya contienen datos, omitiendo inserción inicial';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
TRUNCATE TABLE asientos CASCADE;
TRUNCATE TABLE eventos CASCADE;
