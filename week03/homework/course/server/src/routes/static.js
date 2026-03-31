import Router from '@koa/router';
import { readFileSync, existsSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join, extname, resolve, normalize } from 'path';
import { fail } from '../utils/response.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const DATA_ROOT = resolve(__dirname, '../../data');

const MIME_TYPES = {
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.webp': 'image/webp',
  '.svg': 'image/svg+xml',
};

const router = new Router();

router.get('/(.*)', async (ctx) => {
  const reqPath = ctx.params[0];
  const filePath = normalize(join(DATA_ROOT, reqPath));

  if (!filePath.startsWith(DATA_ROOT)) {
    return fail(ctx, 400, '非法路径');
  }

  const ext = extname(filePath).toLowerCase();
  const mime = MIME_TYPES[ext];
  if (!mime) {
    return fail(ctx, 403, '不支持的文件类型');
  }

  if (!existsSync(filePath)) {
    return fail(ctx, 404, '文件不存在');
  }

  ctx.type = mime;
  ctx.body = readFileSync(filePath);
});

export default router;
