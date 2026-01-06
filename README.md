# Gratitude Journal API

API REST para registrar notas de agradecimiento diarias.

## Requisitos

- Go 1.21 o superior
- GCC (para compilar SQLite)

En Ubuntu/Debian:
```bash
sudo apt install build-essential
```

## Instalacion

1. Clonar el repositorio
2. Copiar el archivo de configuracion:
```bash
cp .env.example .env
```

3. Ejecutar:
```bash
go run cmd/server/main.go
```

O compilar y ejecutar:
```bash
go build -o gratitude-api cmd/server/main.go
./gratitude-api
```

## Configuracion

Variables de entorno en `.env`:

| Variable | Descripcion | Default |
|----------|-------------|---------|
| PORT | Puerto del servidor | 8080 |
| DB_PATH | Ruta de la base de datos | ./db/gratitude.db |

## Endpoints

### Listar notas
```bash
# Todas las notas
curl http://localhost:8080/notes

# Por fecha
curl "http://localhost:8080/notes?date=2026-01-06"

# Por rango de fechas
curl "http://localhost:8080/notes?from=2026-01-01&to=2026-01-31"
```

### Obtener nota por ID
```bash
curl http://localhost:8080/notes/1
```

### Crear nota
```bash
# Con fecha actual
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"content": "Hoy agradezco el buen clima"}'

# Con fecha especifica
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"content": "Agradezco a mi familia", "date": "2026-01-05"}'
```

### Actualizar nota
```bash
curl -X PUT http://localhost:8080/notes/1 \
  -H "Content-Type: application/json" \
  -d '{"content": "Contenido actualizado"}'
```

### Eliminar nota
```bash
curl -X DELETE http://localhost:8080/notes/1
```

## Estructura del proyecto

```
gratitude-journal-api/
├── cmd/server/main.go        # Punto de entrada
├── internal/
│   ├── model/note.go         # Estructura de datos
│   ├── repository/           # Acceso a base de datos
│   ├── service/note.go       # Logica de negocio
│   └── handler/note.go       # Endpoints HTTP
├── db/                       # Base de datos SQLite
├── .env.example              # Plantilla de configuracion
└── go.mod
```

## Arquitectura

Patron Repository con capas:

```
HTTP Request -> Handler -> Service -> Repository -> SQLite
```

- Handler: Recibe peticiones HTTP, parsea JSON
- Service: Logica de negocio, validaciones
- Repository: Operaciones de base de datos
