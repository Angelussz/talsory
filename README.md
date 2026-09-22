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
│   ├── Dockerfile
│   ├── .dockerignore
│   └── .env.example
├── extra_operations/        # Servidor Node.js
│   ├── src/
│   │   ├── index.ts         # Endpoints HTTP
│   │   └── matrix.ts        # max, min, average, isDiagonal, validaciones
│   ├── package.json
│   ├── Dockerfile
│   ├── .dockerignore
│   └── .env.example
├── docker-compose.yml       # Orquestación de ambos contenedores en Docker
├── render.yaml              # Blueprint de despliegue en Render
└── .gitignore
```

## Requisitos

- **Go** 1.27 o superior ([descargar](https://go.dev/dl/))
- **Node.js** 24 o superior (para soporte de `node --env-file`) ([descargar](https://nodejs.org/))
- **Docker** con Docker Compose (para correr con contenedores) ([Docker Desktop](https://www.docker.com/products/docker-desktop/))

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

> **Alternativa:** si prefieres no instalar Go/Node, corre todo con contenedores — ver [Docker / Contenedores](#docker--contenedores).

## Docker / Contenedores

El proyecto incluye un `docker-compose.yml` que levanta **dos contenedores separados**:

- **`extra_operations`** (Node): interno y sin puerto publicado en el host. Incluye un *healthcheck* que verifica que el servidor responda.
- **`factorizacion_qr`** (Go): único servicio publicado en el host, en el puerto `3000`.

Los dos contenedores se comunican por la **red interna de Docker Compose**; Go alcanza a Node usando el nombre del servicio (`http://extra_operations:3001/matrix-stats`). Dentro de un contenedor **no** se usa `localhost` para hablar con otro servicio, porque `localhost` sería el propio contenedor: la URL interna se resuelve por nombre de servicio.

El `docker-compose.yml`:

```yaml
services:
  extra_operations:
    build: ./extra_operations
    environment:
      PORT: 3001
    healthcheck:
      test: ["CMD", "node", "-e", "fetch('http://localhost:3001/').then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"]
      interval: 5s
      timeout: 3s
      retries: 10

  factorizacion_qr:
    build: ./factorizacion_qr
    ports:
      - "3000:3000"
    environment:
      PORT: 3000
      NODE_STATS_URL: http://extra_operations:3001/matrix-stats
    depends_on:
      extra_operations:
        condition: service_healthy
```

- `build: ./...` indica qué Dockerfile usar para cada servicio.
- `environment:` define las variables dentro del contenedor (no usa tu `.env` local).
- `healthcheck`: cada 5s hace un `fetch` al `/` de Node; sin respuesta en 3s cuenta como fallo, y tras 10 intentos el contenedor pasa a `unhealthy`.
- `depends_on` con `condition: service_healthy`: Go **no arranca** hasta que el healthcheck de Node confirme que está respondiendo.

### Cómo correr

Desde la raíz del proyecto (donde está `docker-compose.yml`):

```bash
docker compose up --build   # primera vez, o cuando cambies código/Dockerfiles
docker compose up           # siguientes veces, si no cambió nada
```

| Comando                          | Qué hace                                     |
|----------------------------------|----------------------------------------------|
| `docker compose up --build`      | Construye las imágenes y levanta el stack    |
| `docker compose up`              | Levanta con las imágenes ya construidas      |
| `docker compose down`            | Para y elimina los contenedores              |
| `docker compose logs -f`         | Muestra los logs en vivo                     |
| `docker compose ps`              | Estado de los contenedores y healthchecks    |

> **`--build` o no:** la primera vez siempre es `--build` (aún no hay imágenes). Después, si modificaste `main.go`, `src/` o los `Dockerfile`, vuelve a usar `--build` para que la imagen se reconstruya; si no cambió nada, `docker compose up` alcanza. Usar siempre `--build` no hace daño.

Cuando el stack esté arriba, prueba el flujo completo:

```bash
curl -X POST http://localhost:3000/factorize-qr \
  -H "Content-Type: application/json" \
  -d '{"matrix":[[1,2],[3,4]]}'
```

### Variables de entorno en Docker

En contenedores **no se usa el `.env` local** (ignorado por `.dockerignore`). La configuración vive en el bloque `environment:` del compose. Esto funciona porque tanto `godotenv` (Go) como el flag `--env-file-if-exists` (Node) ignoran la ausencia de `.env` y respetan las variables ya definidas en el entorno del contenedor.

> **Advertencia:** si tienes instancias locales corriendo (`go run .` o `npm run dev`), ciérralas antes de `docker compose up` para evitar conflictos por puertos ya ocupados.

## Despliegue en Render

El proyecto está desplegado en **Render** usando el archivo [`render.yaml`](./render.yaml) (Render Blueprint). Render no lee `docker-compose.yml`, por lo que este archivo define los dos servicios de manera equivalente, construyéndolos con sus Dockerfiles (`runtime: docker`).

```yaml
services:
  - type: web
    name: talsory-extra-operations
    runtime: docker
    repo: https://github.com/Angelussz/talsory
    rootDir: extra_operations
    plan: free
    healthCheckPath: /
    envVars:
      - key: PORT
        value: "3001"

  - type: web
    name: talsory-factorizacion-qr
    runtime: docker
    repo: https://github.com/Angelussz/talsory
    rootDir: factorizacion_qr
    plan: free
    healthCheckPath: /
    envVars:
      - key: PORT
        value: "3000"
      - key: NODE_STATS_URL
        fromService:
          name: talsory-extra-operations
          type: web
          property: host
```

- Cada servicio se construye desde la raíz de su carpeta (`rootDir`) usando su `Dockerfile`.
- `healthCheckPath: /` usa el endpoint raíz de cada servicio como sonda de salud.
- `NODE_STATS_URL` se resuelve automáticamente con `fromService`: Render inyecta el host público de `talsory-extra-operations`; la app Go agrega `/matrix-stats` si falta, así el flujo de stats funciona sin configurar la URL manualmente.

### Cómo se desplegó

1. El repositorio está en GitHub (requisito de Render).
2. En render.com → **New +** → **Blueprint** → conectar GitHub → seleccionar el repo `talsory`.
3. Render detectó `render.yaml` y creó ambos servicios con autodeploy: cada `git push` a la rama principal vuelve a desplegar.
4. URLs resultantes:
   - Go (entrada pública): `https://talsory-factorizacion-qr.onrender.com`
   - Node (stats, interno de la app): `https://talsory-extra-operations.onrender.com`

### Prueba del despliegue

```bash
curl -X POST https://talsory-factorizacion-qr.onrender.com/factorize-qr \
  -H "Content-Type: application/json" \
  -d '{"matrix":[[1,2],[3,4]]}'
```

La respuesta debe incluir `stats` poblado (`max`, `min`, `average`, `qIsDiagonal`, `rIsDiagonal`).

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