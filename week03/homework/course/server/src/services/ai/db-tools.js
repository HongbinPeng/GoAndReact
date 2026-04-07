import db from '../../database/db.js';

export function queryAllCourses() {
  const rows = db.prepare('SELECT * FROM courses ORDER BY created_at DESC').all();
  return { rows, count: rows.length };
}

export function queryAllLearningRecords() {
  const rows = db.prepare(`
    SELECT
      lr.id,
      lr.student_id,
      s.name AS student_name,
      s.student_no,
      lr.course_id,
      c.name AS course_name,
      lr.date,
      lr.duration
    FROM learning_records lr
    LEFT JOIN students s ON s.id = lr.student_id
    LEFT JOIN courses c ON c.id = lr.course_id
    ORDER BY lr.date DESC, lr.id DESC
  `).all();
  return { rows, count: rows.length };
}

export function queryAllStudents() {
  const rows = db.prepare('SELECT * FROM students ORDER BY created_at DESC').all()
    .map((student) => ({
      ...student,
      course_ids: JSON.parse(student.course_ids || '[]'),
    }));
  return { rows, count: rows.length };
}
