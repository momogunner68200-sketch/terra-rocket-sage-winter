export const QTE_KEYS = ["Z", "Q", "S", "D"] as const;
export type QteKey = (typeof QTE_KEYS)[number];

const SIZE = 48;

type Template = { key: QteKey; bits: Uint8Array; weight: number };

let templates: Template[] | null = null;

function ensureTemplates(): Template[] {
  if (templates) return templates;
  const canvas = document.createElement("canvas");
  canvas.width = SIZE;
  canvas.height = SIZE;
  const ctx = canvas.getContext("2d", { willReadFrequently: true });
  if (!ctx) return [];
  templates = QTE_KEYS.map((key) => {
    ctx.fillStyle = "#000";
    ctx.fillRect(0, 0, SIZE, SIZE);
    ctx.fillStyle = "#fff";
    ctx.font = "700 34px sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(key, SIZE / 2, SIZE / 2 + 1);
    const img = ctx.getImageData(0, 0, SIZE, SIZE).data;
    const bits = new Uint8Array(SIZE * SIZE);
    let weight = 0;
    for (let i = 0; i < bits.length; i++) {
      const on = img[i * 4] > 140 ? 1 : 0;
      bits[i] = on;
      weight += on;
    }
    return { key, bits, weight };
  });
  return templates;
}

function otsu(gray: Uint8Array): number {
  const hist = new Uint32Array(256);
  for (const value of gray) hist[value] += 1;
  let sum = 0;
  for (let i = 0; i < 256; i++) sum += i * hist[i];
  let sumB = 0;
  let wB = 0;
  let max = 0;
  let threshold = 128;
  const total = gray.length;
  for (let t = 0; t < 256; t++) {
    wB += hist[t];
    if (!wB) continue;
    const wF = total - wB;
    if (!wF) break;
    sumB += t * hist[t];
    const mB = sumB / wB;
    const mF = (sum - sumB) / wF;
    const between = wB * wF * (mB - mF) ** 2;
    if (between > max) {
      max = between;
      threshold = t;
    }
  }
  return threshold;
}

function normalizeGlyph(mask: Uint8Array, width: number, height: number, on: 0 | 1): Uint8Array | null {
  let minX = width;
  let minY = height;
  let maxX = -1;
  let maxY = -1;
  let count = 0;
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      if (mask[y * width + x] !== on) continue;
      count += 1;
      if (x < minX) minX = x;
      if (y < minY) minY = y;
      if (x > maxX) maxX = x;
      if (y > maxY) maxY = y;
    }
  }
  if (count < 40 || maxX < 0) return null;
  const bw = maxX - minX + 1;
  const bh = maxY - minY + 1;
  if (bw < 10 || bh < 10) return null;
  if (bw > width * 0.9 && bh > height * 0.9) return null;
  const out = new Uint8Array(SIZE * SIZE);
  for (let y = 0; y < SIZE; y++) {
    const sy = minY + Math.min(bh - 1, Math.floor((y / SIZE) * bh));
    for (let x = 0; x < SIZE; x++) {
      const sx = minX + Math.min(bw - 1, Math.floor((x / SIZE) * bw));
      out[y * SIZE + x] = mask[sy * width + sx] === on ? 1 : 0;
    }
  }
  return out;
}

function bestKey(glyph: Uint8Array): QteKey | null {
  const known = ensureTemplates();
  let weight = 0;
  for (const bit of glyph) weight += bit;
  if (weight < 20) return null;
  let best: QteKey | null = null;
  let bestScore = 0;
  let second = 0;
  for (const template of known) {
    let inter = 0;
    for (let i = 0; i < glyph.length; i++) {
      if (glyph[i] && template.bits[i]) inter += 1;
    }
    const score = (2 * inter) / (weight + template.weight);
    if (score > bestScore) {
      second = bestScore;
      bestScore = score;
      best = template.key;
    } else if (score > second) {
      second = score;
    }
  }
  if (!best || bestScore < 0.45 || bestScore - second < 0.06) return null;
  return best;
}

/** Reads only the center of a frame. Nothing is uploaded. */
export function readCenter(
  source: CanvasImageSource,
  sw: number,
  sh: number,
  preview: HTMLCanvasElement | null,
): QteKey | null {
  if (sw < 16 || sh < 16) return null;
  const side = Math.max(16, Math.round(Math.min(sw, sh) * 0.25));
  const sx = Math.round((sw - side) / 2);
  const sy = Math.round((sh - side) / 2);
  const sample = document.createElement("canvas");
  const sampleSize = 96;
  sample.width = sampleSize;
  sample.height = sampleSize;
  const ctx = sample.getContext("2d", { willReadFrequently: true });
  if (!ctx) return null;
  ctx.drawImage(source, sx, sy, side, side, 0, 0, sampleSize, sampleSize);

  if (preview) {
    const pctx = preview.getContext("2d");
    if (pctx) {
      if (preview.width !== 280) preview.width = 280;
      if (preview.height !== 280) preview.height = 280;
      pctx.drawImage(sample, 0, 0, 280, 280);
    }
  }

  const img = ctx.getImageData(0, 0, sampleSize, sampleSize).data;
  const gray = new Uint8Array(sampleSize * sampleSize);
  for (let i = 0; i < gray.length; i++) {
    const offset = i * 4;
    gray[i] = (img[offset] * 0.3 + img[offset + 1] * 0.59 + img[offset + 2] * 0.11) | 0;
  }
  const threshold = otsu(gray);
  const mask = new Uint8Array(gray.length);
  for (let i = 0; i < gray.length; i++) mask[i] = gray[i] > threshold ? 1 : 0;

  const light = normalizeGlyph(mask, sampleSize, sampleSize, 1);
  const dark = normalizeGlyph(mask, sampleSize, sampleSize, 0);
  return (light && bestKey(light)) || (dark && bestKey(dark)) || null;
}

export function drawPractice(canvas: HTMLCanvasElement, key: QteKey) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  canvas.width = 960;
  canvas.height = 540;
  ctx.fillStyle = "#14211c";
  ctx.fillRect(0, 0, 960, 540);
  ctx.strokeStyle = "#0e6b58";
  ctx.lineWidth = 10;
  ctx.beginPath();
  ctx.arc(480, 270, 130, 0, Math.PI * 2);
  ctx.stroke();
  ctx.fillStyle = "#f6f1e7";
  ctx.font = "700 150px sans-serif";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(key, 480, 286);
}

export function nextPrompt(current: QteKey): QteKey {
  const pool = QTE_KEYS.filter((key) => key !== current);
  return pool[Math.floor(Math.random() * pool.length)] ?? "Z";
}
