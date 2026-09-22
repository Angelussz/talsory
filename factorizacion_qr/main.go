package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"gonum.org/v1/gonum/mat"
)

type MatrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type MatrixResponse struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

type MatrixStats struct {
	Max         float64 `json:"max"`
	Min         float64 `json:"min"`
	Average     float64 `json:"average"`
	QIsDiagonal bool    `json:"qIsDiagonal"`
	RIsDiagonal bool    `json:"rIsDiagonal"`
}

type FinalResponse struct {
	Q       [][]float64  `json:"q"`
	R       [][]float64  `json:"r"`
	Stats   *MatrixStats `json:"stats"`
	Warning string       `json:"warning"`
}

// Convierte [][]float64 a un mat.Dense de gonum
func sliceToDense(matrix [][]float64) (*mat.Dense, error) {
	rows := len(matrix)
	if rows == 0 {
		return nil, errors.New("la matriz no puede estar vacía")
	}

	cols := len(matrix[0])
	if cols == 0 {
		return nil, errors.New("las filas no pueden estar vacías")
	}

	// Gonum QR requiere que rows >= cols (m >= n)
	if rows < cols {
		return nil, errors.New("para factorización QR estándar, la cantidad de filas debe ser mayor o igual a las columnas")
	}

	// Aplanar la matriz a un único slice unidimensional
	flatData := make([]float64, 0, rows*cols)
	for _, row := range matrix {
		if len(row) != cols {
			return nil, errors.New("la matriz debe ser rectangular (todas las filas deben tener la misma longitud)")
		}
		flatData = append(flatData, row...)
	}

	return mat.NewDense(rows, cols, flatData), nil
}

// Convierte un mat.Matrix de gonum de regreso a [][]float64 para el JSON
func denseToSlice(m mat.Matrix) [][]float64 {
	rows, cols := m.Dims()
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			result[i][j] = m.At(i, j)
		}
	}
	return result
}

// Obtiene las estadísticas de Q y R desde el servidor de Node.js
func sendToNode(statsURL string, response MatrixResponse) (*MatrixStats, string) {
	payload, err := json.Marshal(response)
	if err != nil {
		return nil, "error al serializar el payload: " + err.Error()
	}

	req, err := http.NewRequest(http.MethodPost, statsURL, bytes.NewReader(payload))
	if err != nil {
		return nil, "error al crear la petición: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "no se pudo conectar con el servidor de stats: " + err.Error()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "error al leer la respuesta del servidor de stats: " + err.Error()
	}

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error string `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &errBody); jsonErr == nil && errBody.Error != "" {
			return nil, "el servidor de stats respondió con estado " + resp.Status + ": " + errBody.Error
		}
		return nil, "el servidor de stats respondió con estado: " + resp.Status
	}

	var stats MatrixStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, "error al decodificar las stats: " + err.Error()
	}

	return &stats, ""
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	statsURL := os.Getenv("NODE_STATS_URL")
	if statsURL == "" {
		statsURL = "http://localhost:3001/matrix-stats"
	}
	if !strings.HasSuffix(statsURL, "/matrix-stats") {
		statsURL += "/matrix-stats"
	}

	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Calculo de factorizacion QR en /factorize-qr")
	})

	app.Post("/factorize-qr", func(c fiber.Ctx) error {
		var req MatrixRequest

		// 1. Parsear el body JSON
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "JSON inválido: " + err.Error(),
			})
		}

		// 2. Convertir y validar matriz rectangular
		A, err := sliceToDense(req.Matrix)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 3. Factorización QR
		var qr mat.QR
		qr.Factorize(A)

		// Extraer Q y R
		var q mat.Dense
		var r mat.Dense
		qr.QTo(&q)
		qr.RTo(&r)

		response := MatrixResponse{
			Q: denseToSlice(&q),
			R: denseToSlice(&r),
		}

		// 4. Enviar Q y R al servidor de Node.js para que calcule las stats
		stats, warning := sendToNode(statsURL, response)

		// 5. Armar respuesta en formato array de arrays
		final := FinalResponse{
			Q:       response.Q,
			R:       response.R,
			Stats:   stats,
			Warning: warning,
		}

		return c.Status(fiber.StatusOK).JSON(final)
	})

	log.Fatal(app.Listen(":" + port))
}
