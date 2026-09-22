import express, { type Request, type Response } from 'express';
const app = express();
const PORT = process.env.PORT || 3001;

app.use(express.json());

app.get('/', (req: Request, res: Response) => {
  res.json({ message: '¡Servidor corriendo con Node.js, TypeScript y Express!' });
});

app.post('/print-qr', (req: Request, res: Response) => {
  const { q, r } = req.body as { q: number[][]; r: number[][] };

  if (!q || !r) {
    res.status(400).json({ error: 'Se esperaban las matrices q y r en el body' });
    return;
  }

  console.log('=== Matrices QR recibidas ===');
  console.log('Matriz Q:');
  console.table(q);
  console.log('Matriz R:');
  console.table(r);

  res.json({ message: 'Matrices QR recibidas y impresas' });
});

// Iniciar servidor
app.listen(PORT, () => {
  console.log(`🚀 Servidor ejecutándose en http://localhost:${PORT}`);
});