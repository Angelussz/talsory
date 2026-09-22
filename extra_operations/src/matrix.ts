export function flattenValues(...matrices: number[][][]): number[] {
  const values = matrices.flat(2);
  if (values.length === 0) {
    throw new Error('Las matrices están vacías');
  }
  return values;
}

export function maxValue(...matrices: number[][][]): number {
  const values = flattenValues(...matrices);
  return Math.max(...values);
}

export function minValue(...matrices: number[][][]): number {
  const values = flattenValues(...matrices);
  return Math.min(...values);
}

export function averageValue(...matrices: number[][][]): number {
  const values = flattenValues(...matrices);
  return values.reduce((sum, v) => sum + v, 0) / values.length;
}

export function isDiagonal(matrix: number[][]): boolean {
  if (matrix.length === 0) return false;
  const n = matrix.length;
  if (!matrix.every((row) => row.length === n)) return false;

  for (let i = 0; i < n; i++) {
    for (let j = 0; j < n; j++) {
      if (i !== j && matrix[i]![j] !== 0) return false;
    }
  }
  return true;
}

export function isValidMatrix(value: unknown): value is number[][] {
  if (!Array.isArray(value) || value.length === 0) return false;
  return value.every(
    (row) =>
      Array.isArray(row) &&
      row.length === (value as number[][]).length &&
      row.every((cell) => typeof cell === 'number' && Number.isFinite(cell)),
  );
}
