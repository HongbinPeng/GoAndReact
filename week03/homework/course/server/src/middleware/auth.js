import jwt from 'jsonwebtoken';

const JWT_SECRET = 'homework_secret_key_2024';

export { JWT_SECRET };

export function authenticateToken(ctx, next) {
  const authHeader = ctx.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1];

  if (!token) {
    ctx.status = 401;
    ctx.body = { code: 401, msg: '未提供认证令牌', data: null };
    return;
  }

  try {
    const user = jwt.verify(token, JWT_SECRET);
    ctx.state.user = user;
    return next();
  } catch {
    ctx.status = 401;
    ctx.body = { code: 401, msg: '令牌无效或已过期', data: null };
  }
}
