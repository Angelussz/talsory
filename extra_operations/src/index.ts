import express, { type Request, type Response } from 'express';
import { averageValue, isDiagonal, isValidMatrix, maxValue, minValue } from './matrix.js';
const app = express();
const PORT = process.env.PORT || 3001;

app.use(express.json());

app.get('/', (req: Request, res: Response) => {
  res.json({ message: '¡Servidor corriendo con Node.js, TypeScript y Express!' });
});


app.post('/matrix-stats', (req: Request, res: Response) => {
  const { q, r } = req.body as { q: unknown; r: unknown };

  if (!isValidMatrix(q) || !isValidMatrix(r)) {
    res.status(400).json({ error: 'Se esperaban las matrices q y r numéricas y no vacías en el body' });
    return;
  }

  try {
    res.json({
      max: maxValue(q, r),
      min: minValue(q, r),
      average: averageValue(q, r),
      qIsDiagonal: isDiagonal(q),
      rIsDiagonal: isDiagonal(r),
    });
  } catch (err) {
    res.status(400).json({ error: err instanceof Error ? err.message : 'Error calculando estadísticas' });
  }
});

// Iniciar servidor
app.listen(PORT, () => {
  console.log(`🚀 Servidor ejecutándose en http://localhost:${PORT}`);
});