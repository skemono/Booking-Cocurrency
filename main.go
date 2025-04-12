package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	_ "github.com/lib/pq"         // Driver de PostgreSQL
	"github.com/pressly/goose/v3" // Herramienta de migración de base de datos
)

const (
	host     = "localhost"
	port     = 5432
	user     = "admin"
	password = "reservasDB123!"
	dbname   = "reservas_db"
)

// reserveAsiento intenta reservar un asiento para un evento dado utilizando una transacción.
// Realiza los siguientes pasos:
// 1. Inicia una transacción.
// 2. Selecciona (y bloquea) el registro del asiento deseado usando FOR UPDATE.
// 3. Verifica que el asiento aún esté disponible ("disponible").
// 4. Actualiza el asiento a "reservado" y luego inserta un registro de reserva.
// 5. Confirma la transacción o realiza un rollback en caso de error.
func reserveAsiento(db *sql.DB, eventoID int, asientoNum int, userID string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("iniciar transacción: %v", err)
	}

	// Bloquea la fila del asiento para prevenir actualizaciones concurrentes.
	var asientoID int
	var estado string
	query := `
	SELECT id, estado
	FROM asientos
	WHERE evento_id = $1 AND numero_asiento = $2
	FOR UPDATE
	`
	err = tx.QueryRow(query, eventoID, asientoNum).Scan(&asientoID, &estado)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("consulta de asiento: %v", err)
	}

	// Si el asiento ya está reservado, realiza un rollback y retorna un error.
	if estado != "disponible" {
		tx.Rollback()
		return fmt.Errorf("el asiento %d no está disponible", asientoNum)
	}

	// Actualiza el asiento para marcarlo como reservado.
	_, err = tx.Exec("UPDATE asientos SET estado = 'reservado' WHERE id = $1", asientoID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("actualizar asiento: %v", err)
	}

	// Inserta el registro de la reserva.
	_, err = tx.Exec(
		"INSERT INTO reservas (usuario_id, asiento_id, estado_reserva) VALUES ($1, $2, 'exitosa')",
		userID, asientoID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insertar reserva: %v", err)
	}

	// Confirma la transacción.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("confirmar transacción: %v", err)
	}

	return nil
}

func main() {
	// Construye la cadena de conexión a PostgreSQL.
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	// Abre una conexión a la base de datos.
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}
	defer db.Close()

	// Verifica la conexión.
	if err = db.Ping(); err != nil {
		log.Fatalf("Error al verificar la conexión: %v", err)
	}

	// -------------------------------
	// Ejecutar migraciones con Goose
	// -------------------------------
	// Configura el dialecto como "postgres" para Goose.
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error al configurar el dialecto: %v", err)
	}

	// Ejecuta las migraciones desde la carpeta "./db".
	if err := goose.Up(db, "./db"); err != nil {
		log.Fatalf("Error al ejecutar las migraciones: %v", err)
	}
	fmt.Println("¡Migraciones aplicadas exitosamente!")

	// -------------------------------
	// Simular reservas concurrentes
	// -------------------------------
	// Simulamos intentos de reserva para el evento con ID 1,
	// que  es "Concierto de Orquesta Sinfónica" con 50 asientos.
	eventoID := 1
	totalSeats := 50

	// Usamos un WaitGroup para rastrear la finalización de todas las gorutinas.
	var wg sync.WaitGroup
	const numReservations = 20 // Número de intentos de reserva concurrentes

	// Creamos un generador de números aleatorios para seleccionar asientos al azar.
	randSource := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(randSource)

	// Inicia intentos de reserva concurrentes.
	for i := 0; i < numReservations; i++ {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()

			// Selecciona un número de asiento aleatorio entre 1 y totalSeats.
			asientoNum := rnd.Intn(totalSeats) + 1

			// Intenta reservar el asiento seleccionado.
			err := reserveAsiento(db, eventoID, asientoNum, userID)
			if err != nil {
				log.Printf("Reserva fallida para el usuario %s en el asiento %d: %v", userID, asientoNum, err)
			} else {
				log.Printf("Reserva exitosa para el usuario %s en el asiento %d", userID, asientoNum)
			}
		}(fmt.Sprintf("user_%d", i+1))
	}

	wg.Wait()
	fmt.Println("Se completaron todos los intentos de reserva.")
}
