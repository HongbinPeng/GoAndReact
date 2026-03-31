import Koa from 'koa';
import Router from '@koa/router';
import cors from '@koa/cors';
import bodyParser from 'koa-bodyparser';
import { existsSync, readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, extname, join } from 'path';
import { initDatabase } from './database/init.js';
import authRoutes from './routes/auth.js';
import dashboardRoutes from './routes/dashboard.js';
import courseRoutes from './routes/courses.js';
import studentRoutes from './routes/students.js';
import summaryRoutes from './routes/summary.js';
import staticRoutes from './routes/static.js';

const app = new Koa();
const router = new Router();
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const CLIENT_DIST = join(__dirname, '../../client/dist');

initDatabase();

app.use(cors({ credentials: true }));
app.use(bodyParser());

app.use(async (ctx, next) => {
  try {
    await next();
  } catch (err) {
    const status = err.status || 500;
    ctx.status = status;
    ctx.body = { code: status, msg: err.message || '服务器内部错误', data: null };
    console.error(`[${new Date().toISOString()}] ${err.message}`);
  }
});

router.use('/api/auth', authRoutes.routes());
router.use('/api/dashboard', dashboardRoutes.routes());
router.use('/api/courses', courseRoutes.routes());
router.use('/api/students', studentRoutes.routes());
router.use('/api/summary', summaryRoutes.routes());
router.use('/api/static', staticRoutes.routes());

app.use(async (ctx, next) => {
  if (ctx.path.startsWith('/api')) {
    return next();
  }

  const reqPath = ctx.path === '/' ? '/index.html' : ctx.path;
  const filePath = join(CLIENT_DIST, reqPath);
  const ext = extname(filePath);

  if (existsSync(filePath) && ext) {
    ctx.type = ext;
    ctx.body = readFileSync(filePath);
    return;
  }

  const indexPath = join(CLIENT_DIST, 'index.html');
  if (existsSync(indexPath)) {
    ctx.type = 'html';
    ctx.body = readFileSync(indexPath, 'utf-8');
    return;
  }

  return next();
});

app.use(router.routes());
app.use(router.allowedMethods());

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`服务端已启动: http://localhost:${PORT}`);
});
