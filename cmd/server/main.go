package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"gratitude-journal-api/internal/handler"
	"gratitude-journal-api/internal/repository"
	"gratitude-journal-api/internal/service"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando valores por defecto")
	}

	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./db/gratitude.db")

	// Inicializar base de datos
	db, err := repository.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Error al conectar con la base de datos: %v", err)
	}
	defer db.Close()

	// Inicializar capas: Repository → Service → Handler
	noteRepo := repository.NewNoteRepository(db)
	noteService := service.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteService)

	// Registrar rutas
	mux := http.NewServeMux()
	mux.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			noteHandler.GetAll(w, r)
		case http.MethodPost:
			noteHandler.Create(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/notes/", func(w http.ResponseWriter, r *http.Request) {
		// Verificar que hay un ID en la ruta
		path := strings.TrimPrefix(r.URL.Path, "/notes/")
		if path == "" {
			http.Error(w, "ID requerido", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			noteHandler.GetByID(w, r)
		case http.MethodPut:
			noteHandler.Update(w, r)
		case http.MethodDelete:
			noteHandler.Delete(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
	})

	// Iniciar servidor
	log.Printf("Servidor iniciado en http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
