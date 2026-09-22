# Talsory

Sistema que calcula la factorización QR de una matriz mediante un backend en Go y un servicio de operaciones adicionales (Node.js).

## Estructura del proyecto

```
talsory/
├── factorizacion_qr/     # Backend en Go: factorización QR
└── extra_operations/     # Servidor Node.js (a desarrollar)
```

## factorizacion_qr

Backend en Go creado con Fiber v3 que expone un endpoint para calcular la factorización QR de una matriz.

### Cómo correrlo

```bash
cd factorizacion_qr
go run main.go
```

El servidor queda escuchando en `http://localhost:3000`.

### Endpoints

| Método | Ruta            | Descripción                                        |
|--------|-----------------|----------------------------------------------------|
| GET    | `/`             | Texto de verificación                              |
| POST   | `/factorize-qr` | Calcula la factorización QR de una matriz enviada  |

`POST /factorize-qr` recibe:

```json
{
  "matrix": [[1, 2], [3, 4], [5, 6]]
}
```

Y responde:

```json
{
  "q": [[...], [...]],
  "r": [[...], [...]]
}
```

### Validaciones

- La matriz no puede estar vacía.
- Todas las filas deben tener la misma longitud (matriz rectangular).
- La cantidad de filas debe ser mayor o igual a la cantidad de columnas.
- JSON inválido devuelve error 400.

### Factorización

La factorización QR se realiza con la librería [gonum](https://gonum.org/), extrayendo las matrices Q y R del resultado.

## extra_operations

Carpeta reservada para el servidor Node.js con las operaciones adicionales (aún sin contenido).