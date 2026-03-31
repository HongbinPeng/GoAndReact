import Router from '@koa/router';
import jwt from 'jsonwebtoken';
import bcrypt from 'bcryptjs';
import db from '../database/db.js';
import { JWT_SECRET, authenticateToken } from '../middleware/auth.js';
import { success, fail } from '../utils/response.js';

const router = new Router();

router.post('/login', async (ctx) => {
  const { username, password } = ctx.request.body;

  if (!username || !password) {
    return fail(ctx, 400, '请输入用户名和密码');
  }

  const user = db.prepare('SELECT * FROM users WHERE username = ?').get(username);
  if (!user) {
    return fail(ctx, 401, '用户名或密码错误');
  }

  const valid = bcrypt.compareSync(password, user.password);
  if (!valid) {
    return fail(ctx, 401, '用户名或密码错误');
  }

  const token = jwt.sign(
    { id: user.id, username: user.username, role: user.role, name: user.name },
    JWT_SECRET,
    { expiresIn: '7d' }
  );

  const { password: _, ...userInfo } = user;
  success(ctx, { token, user: userInfo });
});

router.get('/me', authenticateToken, async (ctx) => {
  const user = db.prepare('SELECT id, username, name, role, avatar, created_at FROM users WHERE id = ?').get(ctx.state.user.id);
  if (!user) {
    return fail(ctx, 404, '用户不存在');
  }
  success(ctx, user);
});

export default router;
