# Talsory

Sistema que calcula la **factorización QR** de una matriz mediante un backend en **Go** y obtiene estadísticas adicionales (máximo, mínimo, promedio y verificación de diagonal) desde un servicio en **Node.js**.

## Flujo

1. El cliente envía una matriz a `POST /factorize-qr` (servidor **Go**, Fiber v3).
2. Go valida la matriz y calcula la factorización QR usando la librería [gonum](https://gonum.org/), obteniendo las matrices **Q** y **R**.
3. Go envía esas matrices a `POST /matrix-stats` (servidor **Node.js**, Express) para que calcule las estadísticas.
4. Node devuelve `max`, `min`, `average`, `qIsDiagonal` y `rIsDiagonal`.
5. Go fusiona todo y responde al cliente: **Q + R + stats**.

```
Cliente ── POST /factorize-qr { matrix } ──> Go (Fiber + gonum)
                                                    │ factoriza QR → Q, R
                                                    ▼
                              POST /matrix-stats { q, r } ──> Node (Express)
                                                    ▲              │ max, min, average,
                                                    │              │ qIsDiagonal, rIsDiagonal
                                                    └──────────────┘
Cliente ── { q, r, stats, warning } <─────────── Go (fusiona respuestas)
```

> **Comportamiento degradado:** si el servidor de Node está apagado o no responde, Go responde igual con `q` y `r`, pero `stats` viene como `null` y `warning` contiene el error devuelto por el servicio de stats (o el error de conexión). Go nunca deja de responder por una falla en Node.

## Estructura del proyecto

```
talsory/
├── factorizacion_qr/        # Backend en Go
│   ├── main.go              # Endpoint /factorize-qr + integración con Node
│   ├── go.mod / go.sum
│   └── .env.example
├── extra_operations/        # Servidor Node.js
│   ├── src/
│   │   ├── index.ts         # Endpoints HTTP
│   │   └── matrix.ts        # max, min, average, isDiagonal, validaciones
│   ├── package.json
│   └── .env.example
└── .gitignore
```

## Requisitos

- **Go** 1.27 o superior ([descargar](https://go.dev/dl/))
- **Node.js** 24 o superior (para soporte de `node --env-file`) ([descargar](https://nodejs.org/))

## Instalación (proyecto clonado)

1. Clonar el repositorio y entrar:

   ```bash
   git clone 
   cd talsory
   ```

2. Configurar el servidor Node (se inicia primero):

   ```bash
   cd extra_operations
   npm install
   cp .env.example .env
   ```

   Variables disponibles (`extra_operations/.env`):

   | Variable | Descripción        | Default |
   |----------|--------------------|---------|
   | `PORT`   | Puerto del servidor | `3001`  |

3. Configurar el servidor Go:

   ```bash
   cd ../factorizacion_qr
   cp .env.example .env
   ```

   Variables disponibles (`factorizacion_qr/.env`):

   | Variable          | Descripción                             | Default                           |
   |-------------------|-----------------------------------------|-----------------------------------|
   | `PORT`            | Puerto del servidor Go                  | `3000`                            |
   | `NODE_STATS_URL`  | URL del endpoint `/matrix-stats` de Node | `http://localhost:3001/matrix-stats` |

## Cómo correr

> **Importante:** arranca primero el servidor de Node, después el de Go, para que el flujo de stats funcione completo.

1. Servidor Node (en `extra_operations/`):

   ```bash
   npm run dev
   ```

   Escucha en `http://localhost:3001`. Para producción: `npm run build` y luego `npm start`.

2. Servidor Go (en `factorizacion_qr/`):

   ```bash
   go run .
   ```

   Escucha en `http://localhost:3000` (o el valor de `PORT`).

## Endpoints

### Go — `factorizacion_qr`

| Método | Ruta            | Descripción                                        |
|--------|-----------------|----------------------------------------------------|
| GET    | `/`             | Texto de verificación                              |
| POST   | `/factorize-qr` | Calcula la factorización QR y las stats de la matriz |

`POST /factorize-qr` recibe:

```json
{
  "matrix": [[1, 2], [3, 4]]
}
```

Y responde:

```json
{
  "q": [
    [-0.316227766016838, -0.9486832980505138],
    [-0.9486832980505138, 0.316227766016838]
  ],
  "r": [
    [-3.1622776601683795, -4.427188724235731],
    [0, -0.6324555320336751]
  ],
  "stats": {
    "max": 0.316227766016838,
    "min": -4.427188724235731,
    "average": -1.2649110640673516,
    "qIsDiagonal": false,
    "rIsDiagonal": false
  },
  "warning": ""
}
```

Validaciones:

- La matriz no puede estar vacía.
- Todas las filas deben tener la misma longitud (matriz rectangular).
- La cantidad de filas debe ser mayor o igual a la cantidad de columnas.
- JSON inválido devuelve error `400`.

### Node — `extra_operations`

| Método | Ruta            | Descripción                                        |
|--------|-----------------|----------------------------------------------------|
| GET    | `/`             | Mensaje de verificación                            |
| POST   | `/matrix-stats` | Calcula stats (max, min, average, diagonal) de q y r |

`POST /matrix-stats` recibe:

```json
{
  "q": [[1, 0], [0, 2]],
  "r": [[3, 4], [5, 6]]
}
```

Y responde:

```json
{
  "max": 6,
  "min": 0,
  "average": 2.625,
  "qIsDiagonal": true,
  "rIsDiagonal": false
}
```

- `max`: valor máximo entre todas las matrices.
- `min`: valor mínimo entre todas las matrices.
- `average`: promedio de todos los valores.
- `qIsDiagonal` / `rIsDiagonal`: `true` si la matriz es cuadrada y todos sus elementos fuera de la diagonal principal son `0`.

Validaciones:

- Las matrices `q` y `r` deben existir en el body, ser numéricas, no vacías y cuadradas-rectangulares; si no, error `400`.