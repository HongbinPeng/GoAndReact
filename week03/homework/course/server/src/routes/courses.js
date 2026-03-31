import Router from '@koa/router';
import db from '../database/db.js';
import { authenticateToken } from '../middleware/auth.js';
import { success, fail } from '../utils/response.js';

const router = new Router();

router.get('/', authenticateToken, async (ctx) => {
  const { keyword = '', status = '', category = '', page = 1, pageSize = 10, sortField = '', sortOrder = '' } = ctx.query;
  const offset = (Number(page) - 1) * Number(pageSize);

  let where = 'WHERE 1=1';
  const params = [];

  if (keyword) {
    where += ' AND (name LIKE ? OR instructor LIKE ?)';
    params.push(`%${keyword}%`, `%${keyword}%`);
  }
  if (status) {
    where += ' AND status = ?';
    params.push(status);
  }
  if (category) {
    where += ' AND category = ?';
    params.push(category);
  }

  const allowedSortFields = ['student_count', 'lesson_count', 'created_at', 'name'];
  let orderBy = 'ORDER BY created_at DESC';
  if (sortField && allowedSortFields.includes(sortField) && ['ascend', 'descend'].includes(sortOrder)) {
    const dir = sortOrder === 'ascend' ? 'ASC' : 'DESC';
    orderBy = `ORDER BY ${sortField} ${dir}`;
  }

  const total = db.prepare(`SELECT COUNT(*) as count FROM courses ${where}`).get(...params).count;
  const list = db.prepare(`SELECT * FROM courses ${where} ${orderBy} LIMIT ? OFFSET ?`)
    .all(...params, Number(pageSize), offset);

  success(ctx, { list, total, page: Number(page), pageSize: Number(pageSize) });
});

router.get('/categories', authenticateToken, async (ctx) => {
  const categories = db.prepare("SELECT DISTINCT category FROM courses WHERE category != '' ORDER BY category")
    .all()
    .map(r => r.category);
  success(ctx, categories);
});

router.get('/:id', authenticateToken, async (ctx) => {
  const course = db.prepare('SELECT * FROM courses WHERE id = ?').get(ctx.params.id);
  if (!course) {
    return fail(ctx, 404, '课程不存在');
  }
  success(ctx, course);
});

router.post('/', authenticateToken, async (ctx) => {
  const { name, description, instructor, category, status, lesson_count } = ctx.request.body;

  if (!name) {
    return fail(ctx, 400, '课程名称不能为空');
  }

  const result = db.prepare(`
    INSERT INTO courses (name, description, instructor, category, status, lesson_count)
    VALUES (?, ?, ?, ?, ?, ?)
  `).run(name, description || '', instructor || '', category || '', status || 'draft', lesson_count || 0);

  const course = db.prepare('SELECT * FROM courses WHERE id = ?').get(result.lastInsertRowid);
  ctx.status = 201;
  success(ctx, course);
});

router.put('/:id', authenticateToken, async (ctx) => {
  const existing = db.prepare('SELECT * FROM courses WHERE id = ?').get(ctx.params.id);
  if (!existing) {
    return fail(ctx, 404, '课程不存在');
  }

  const { name, description, instructor, category, status, lesson_count } = ctx.request.body;

  db.prepare(`
    UPDATE courses SET name = ?, description = ?, instructor = ?, category = ?, status = ?, lesson_count = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
  `).run(
    name ?? existing.name,
    description ?? existing.description,
    instructor ?? existing.instructor,
    category ?? existing.category,
    status ?? existing.status,
    lesson_count ?? existing.lesson_count,
    ctx.params.id
  );

  const course = db.prepare('SELECT * FROM courses WHERE id = ?').get(ctx.params.id);
  success(ctx, course);
});

router.delete('/:id', authenticateToken, async (ctx) => {
  const existing = db.prepare('SELECT * FROM courses WHERE id = ?').get(ctx.params.id);
  if (!existing) {
    return fail(ctx, 404, '课程不存在');
  }

  db.prepare('DELETE FROM courses WHERE id = ?').run(ctx.params.id);
  success(ctx, null, '删除成功');
});

router.patch('/:id/status', authenticateToken, async (ctx) => {
  const existing = db.prepare('SELECT * FROM courses WHERE id = ?').get(ctx.params.id);
  if (!existing) {
    return fail(ctx, 404, '课程不存在');
  }

  const newStatus = existing.status === 'published' ? 'draft' : 'published';
  db.prepare('UPDATE courses SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?')
    .run(newStatus, ctx.params.id);

  const course = db.prepare('SELECT * FROM courses WHERE id = ?').get(ctx.params.id);
  success(ctx, course);
});

export default router;
