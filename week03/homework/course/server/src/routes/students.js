import Router from '@koa/router';
import db from '../database/db.js';
import { authenticateToken } from '../middleware/auth.js';
import { success, fail } from '../utils/response.js';

const router = new Router();

router.get('/', authenticateToken, async (ctx) => {
  const { keyword = '', className = '', status = '', courseId = '', page = 1, pageSize = 10 } = ctx.query;
  const offset = (Number(page) - 1) * Number(pageSize);

  let where = 'WHERE 1=1';
  const params = [];

  if (keyword) {
    where += ' AND (name LIKE ? OR student_no LIKE ?)';
    params.push(`%${keyword}%`, `%${keyword}%`);
  }
  if (className) {
    where += ' AND class_name = ?';
    params.push(className);
  }
  if (status) {
    where += ' AND status = ?';
    params.push(status);
  }

  let rows = db.prepare(`SELECT * FROM students ${where} ORDER BY created_at DESC`).all(...params);

  if (courseId) {
    rows = rows.filter(s => {
      const ids = JSON.parse(s.course_ids || '[]');
      return ids.includes(Number(courseId));
    });
  }

  const total = rows.length;
  const list = rows.slice(offset, offset + Number(pageSize)).map(s => ({
    ...s,
    course_ids: JSON.parse(s.course_ids || '[]'),
  }));

  success(ctx, { list, total, page: Number(page), pageSize: Number(pageSize) });
});

router.get('/classes', authenticateToken, async (ctx) => {
  const classes = db.prepare("SELECT DISTINCT class_name FROM students WHERE class_name != '' ORDER BY class_name")
    .all()
    .map(r => r.class_name);
  success(ctx, classes);
});

router.get('/:id', authenticateToken, async (ctx) => {
  const student = db.prepare('SELECT * FROM students WHERE id = ?').get(ctx.params.id);
  if (!student) {
    return fail(ctx, 404, '学生不存在');
  }
  student.course_ids = JSON.parse(student.course_ids || '[]');

  const courses = db.prepare('SELECT id, name FROM courses').all();
  const enrolledCourses = courses.filter(c => student.course_ids.includes(c.id));

  success(ctx, { ...student, enrolledCourses });
});

// TODO: 实现创建学生接口 POST /
// 提示：参考 API.md 中 4.4 的接口说明，以及 courses.js 中创建课程的实现
// 需要：参数校验、学号唯一性检查、插入数据库、更新选课计数
router.post('/', authenticateToken, async (ctx) => {
  const { name, student_no, class_name, phone, email, status, course_ids } = ctx.request.body;

  if (!name || !student_no) {
    return fail(ctx, 400, '学生姓名和学号不能为空');
  }

  const duplicated = db.prepare('SELECT id FROM students WHERE student_no = ?').get(student_no);
  if (duplicated) {
    return fail(ctx, 400, '学号已存在');
  }

  const normalizedCourseIds = Array.isArray(course_ids)
    ? course_ids.filter((id) => Number.isInteger(id))
    : [];

  const result = db.prepare(`
    INSERT INTO students (name, student_no, class_name, phone, email, status, course_ids)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(
    name,
    student_no,
    class_name || '',
    phone || '',
    email || '',
    status || 'active',
    JSON.stringify(normalizedCourseIds)
  );

  updateCourseCounts();

  const student = db.prepare('SELECT * FROM students WHERE id = ?').get(result.lastInsertRowid);
  student.course_ids = JSON.parse(student.course_ids || '[]');
  ctx.status = 201;
  success(ctx, student);
});

// TODO: 实现更新学生接口 PUT /:id
// 提示：参考 API.md 中 4.5 的接口说明，以及 courses.js 中更新课程的实现
// 需要：检查学生是否存在、学号唯一性校验、更新数据库、更新选课计数
router.put('/:id', authenticateToken, async (ctx) => {
  const existing = db.prepare('SELECT * FROM students WHERE id = ?').get(ctx.params.id);
  if (!existing) {
    return fail(ctx, 404, '学生不存在');
  }

  const { name, student_no, class_name, phone, email, status, course_ids } = ctx.request.body;

  if (student_no && student_no !== existing.student_no) {
    const duplicated = db.prepare('SELECT id FROM students WHERE student_no = ? AND id != ?').get(student_no, ctx.params.id);
    if (duplicated) {
      return fail(ctx, 400, '学号已存在');
    }
  }

  const nextCourseIds = Array.isArray(course_ids)
    ? course_ids.filter((id) => Number.isInteger(id))
    : JSON.parse(existing.course_ids || '[]');

  db.prepare(`
    UPDATE students
    SET name = ?, student_no = ?, class_name = ?, phone = ?, email = ?, status = ?, course_ids = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
  `).run(
    name ?? existing.name,
    student_no ?? existing.student_no,
    class_name ?? existing.class_name,
    phone ?? existing.phone,
    email ?? existing.email,
    status ?? existing.status,
    JSON.stringify(nextCourseIds),
    ctx.params.id
  );

  updateCourseCounts();

  const student = db.prepare('SELECT * FROM students WHERE id = ?').get(ctx.params.id);
  student.course_ids = JSON.parse(student.course_ids || '[]');
  success(ctx, student);
});

// TODO: 实现删除学生接口 DELETE /:id
// 提示：参考 API.md 中 4.6 的接口说明，以及 courses.js 中删除课程的实现
// 需要：检查学生是否存在、删除数据库记录、更新选课计数
router.delete('/:id', authenticateToken, async (ctx) => {
  const existing = db.prepare('SELECT * FROM students WHERE id = ?').get(ctx.params.id);
  if (!existing) {
    return fail(ctx, 404, '学生不存在');
  }

  db.prepare('DELETE FROM students WHERE id = ?').run(ctx.params.id);
  updateCourseCounts();
  success(ctx, null, '删除成功');
});

function updateCourseCounts() {
  const courses = db.prepare('SELECT id FROM courses').all();
  const students = db.prepare('SELECT course_ids FROM students').all();

  for (const course of courses) {
    const count = students.filter(s => {
      const ids = JSON.parse(s.course_ids || '[]');
      return ids.includes(course.id);
    }).length;
    db.prepare('UPDATE courses SET student_count = ? WHERE id = ?').run(count, course.id);
  }
}

export default router;
