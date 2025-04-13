package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"         // Driver de PostgreSQL
	"github.com/pressly/goose/v3" // Herramienta de migración de base de datos
)

const (
	host       = "localhost"
	port       = 5432
	user       = "admin"
	password   = "reservasDB123!"
	dbname     = "reservas_db"
	totalSeats = 50 // Total de asientos disponibles para el evento.
)

// reservationResult almacena el resultado de cada intento de reserva.
type reservationResult struct {
	success  bool
	duration time.Duration
}

// testResult contiene los resultados agregados de cada prueba.
type testResult struct {
	numUsuarios   int
	isolationName string
	successCount  int
	failureCount  int
	avgDuration   time.Duration
}

// reserveAsiento intenta reservar un asiento para un evento dado utilizando una transacción
// con el nivel de aislamiento especificado.
func reserveAsiento(db *sql.DB, eventoID int, asientoNum int, userID string, isolation sql.IsolationLevel) error {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: isolation})
	if err != nil {
		return fmt.Errorf("iniciar transacción: %v", err)
	}

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

	// Si el asiento ya está reservado se cancela la transacción.
	if estado != "disponible" {
		tx.Rollback()
		return fmt.Errorf("el asiento %d no está disponible", asientoNum)
	}

	// Actualiza el asiento a "reservado".
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

// resetAllSeats restablece el estado de todos los asientos del evento a "disponible"
// y elimina cualquier reserva previa asociada.
func resetAllSeats(db *sql.DB, eventoID int) error {
	_, err := db.Exec("UPDATE asientos SET estado = 'disponible' WHERE evento_id = $1", eventoID)
	if err != nil {
		return fmt.Errorf("resetear asientos: %v", err)
	}

	_, err = db.Exec("DELETE FROM reservas WHERE asiento_id IN (SELECT id FROM asientos WHERE evento_id = $1)", eventoID)
	if err != nil {
		return fmt.Errorf("eliminar reservas: %v", err)
	}

	return nil
}

// runTest ejecuta una prueba de reservas concurrentes con numUsuarios intentos utilizando el
// nivel de aislamiento especificado y retorna un resumen.
func runTest(db *sql.DB, numUsuarios int, isolation sql.IsolationLevel, isolationName string, eventoID int, totalSeats int) testResult {
	// Restablece el estado de todos los asientos del evento.
	if err := resetAllSeats(db, eventoID); err != nil {
		log.Fatalf("Error al resetear asientos del evento: %v", err)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan reservationResult, numUsuarios)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < numUsuarios; i++ {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			start := time.Now()
			// Selecciona un asiento al azar.
			asientoNum := rnd.Intn(totalSeats) + 1
			err := reserveAsiento(db, eventoID, asientoNum, userID, isolation)
			duration := time.Since(start)
			success := err == nil

			resultsChan <- reservationResult{
				success:  success,
				duration: duration,
			}

			if !success {
				log.Printf("Reserva fallida para %s en asiento %d: %v", userID, asientoNum, err)
			} else {
				log.Printf("Reserva exitosa para %s en asiento %d", userID, asientoNum)
			}
		}(fmt.Sprintf("user_%d", i+1))
	}

	wg.Wait()
	close(resultsChan)

	var successCount, failureCount int
	var totalDuration time.Duration
	count := 0

	for r := range resultsChan {
		count++
		totalDuration += r.duration
		if r.success {
			successCount++
		} else {
			failureCount++
		}
	}

	avgDuration := time.Duration(0)
	if count > 0 {
		avgDuration = totalDuration / time.Duration(count)
	}

	return testResult{
		numUsuarios:   numUsuarios,
		isolationName: isolationName,
		successCount:  successCount,
		failureCount:  failureCount,
		avgDuration:   avgDuration,
	}
}

// exportToCSV genera un archivo CSV con la información consolidada de las pruebas.
func exportToCSV(results []testResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error al crear el archivo CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Escribir encabezado.
	header := []string{"Usuarios", "Nivel de Aislamiento", "Reservas Exitosas", "Reservas Fallidas", "Tiempo Promedio (ms)"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error al escribir el encabezado CSV: %v", err)
	}

	// Escribir cada resultado.
	for _, r := range results {
		record := []string{
			fmt.Sprintf("%d", r.numUsuarios),
			r.isolationName,
			fmt.Sprintf("%d", r.successCount),
			fmt.Sprintf("%d", r.failureCount),
			fmt.Sprintf("%d", r.avgDuration.Milliseconds()),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error al escribir registro CSV: %v", err)
		}
	}

	return nil
}

func main() {
	// Construye la cadena de conexión a PostgreSQL.
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Error al verificar la conexión: %v", err)
	}

	// Ejecución de migraciones con Goose.
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error al configurar el dialecto: %v", err)
	}
	if err := goose.Up(db, "./db"); err != nil {
		log.Fatalf("Error al ejecutar las migraciones: %v", err)
	}
	fmt.Println("¡Migraciones aplicadas exitosamente!")

	// Parámetros del evento.
	eventoID := 1

	// Definición de los casos de prueba.
	testCases := []struct {
		numUsuarios   int
		isolationName string
		isolation     sql.IsolationLevel
	}{
		{numUsuarios: 5, isolationName: "READ COMMITTED", isolation: sql.LevelReadCommitted},
		{numUsuarios: 10, isolationName: "REPEATABLE READ", isolation: sql.LevelRepeatableRead},
		{numUsuarios: 20, isolationName: "SERIALIZABLE", isolation: sql.LevelSerializable},
		{numUsuarios: 30, isolationName: "SERIALIZABLE", isolation: sql.LevelSerializable},
	}

	// Imprime la cabecera de la tabla de resultados en consola.
	fmt.Println("\nUsuarios  Nivel de Aislamiento  Reservas Exitosas  Reservas Fallidas  Tiempo Promedio (ms)")
	fmt.Println("-------------------------------------------------------------------------------")

	// Almacena los resultados en un slice para luego exportarlos a CSV.
	var allResults []testResult
	for _, tc := range testCases {
		result := runTest(db, tc.numUsuarios, tc.isolation, tc.isolationName, eventoID, totalSeats)
		allResults = append(allResults, result)
		fmt.Printf("%-9d %-22s %-19d %-18d %-10d\n", result.numUsuarios, result.isolationName, result.successCount, result.failureCount, result.avgDuration.Milliseconds())
		time.Sleep(1 * time.Second)
	}

	// Exporta los resultados a un archivo CSV.
	csvFilename := "resultados.csv"
	if err := exportToCSV(allResults, csvFilename); err != nil {
		log.Fatalf("Error al exportar resultados a CSV: %v", err)
	}
	fmt.Printf("\nArchivo CSV generado exitosamente: %s\n", csvFilename)
}
