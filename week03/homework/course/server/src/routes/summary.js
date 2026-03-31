import Router from '@koa/router';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { authenticateToken } from '../middleware/auth.js';
import { success, fail } from '../utils/response.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const router = new Router();

router.get('/', authenticateToken, async (ctx) => {
  try {
    const mdPath = join(__dirname, '../../data/summary.md');
    const content = readFileSync(mdPath, 'utf-8');
    success(ctx, { content });
  } catch {
    fail(ctx, 500, '读取学习总结失败');
  }
});

export default router;
