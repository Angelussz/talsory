package main

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
	"gonum.org/v1/gonum/mat"
)

type MatrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type MatrixResponse struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
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

func main() {
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

		// 4. Armar respuesta en formato array de arrays
		response := MatrixResponse{
			Q: denseToSlice(&q),
			R: denseToSlice(&r),
		}

		// Nota: Si vas a enviar estos datos a la API de Node.js,
		// aquí es donde harías la llamada HTTP a Node usando http.Post()
		// o el cliente HTTP que prefieras.

		return c.Status(fiber.StatusOK).JSON(response)
	})

	log.Fatal(app.Listen(":3000"))
}
