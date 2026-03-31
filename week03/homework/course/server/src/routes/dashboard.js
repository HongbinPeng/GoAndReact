import Router from '@koa/router';
import db from '../database/db.js';
import { authenticateToken } from '../middleware/auth.js';
import { success } from '../utils/response.js';

const router = new Router();

router.get('/', authenticateToken, async (ctx) => {
  const totalCourses = db.prepare('SELECT COUNT(*) as count FROM courses').get().count;
  const publishedCourses = db.prepare("SELECT COUNT(*) as count FROM courses WHERE status = 'published'").get().count;
  const totalStudents = db.prepare('SELECT COUNT(*) as count FROM students').get().count;
  const activeStudents = db.prepare("SELECT COUNT(*) as count FROM students WHERE status = 'active'").get().count;

  const enrollment = db.prepare(`
    SELECT c.name, c.student_count as value
    FROM courses c
    WHERE c.status = 'published'
    ORDER BY c.student_count DESC
    LIMIT 8
  `).all();

  const today = new Date();
  const activity = [];
  for (let i = 6; i >= 0; i--) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);
    const dateStr = date.toISOString().split('T')[0];

    const record = db.prepare(`
      SELECT COUNT(DISTINCT student_id) as students, COALESCE(SUM(duration), 0) as duration
      FROM learning_records WHERE date = ?
    `).get(dateStr);

    const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
    activity.push({
      date: dateStr,
      label: weekDays[date.getDay()],
      students: record.students,
      duration: Math.round(record.duration / 60),
    });
  }

  const statusDist = [
    { name: '活跃学生', value: activeStudents },
    { name: '非活跃学生', value: totalStudents - activeStudents },
  ];

  const categoryDist = db.prepare(`
    SELECT category as name, COUNT(*) as value FROM courses
    WHERE category != '' GROUP BY category ORDER BY value DESC
  `).all();

  success(ctx, {
    stats: {
      totalCourses,
      publishedCourses,
      totalStudents,
      activeStudents,
    },
    charts: {
      enrollment,
      activity,
      statusDist,
      categoryDist,
    },
  });
});

export default router;
