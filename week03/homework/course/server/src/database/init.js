import db from './db.js';
import bcrypt from 'bcryptjs';

export function initDatabase() {
  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      username TEXT UNIQUE NOT NULL,
      password TEXT NOT NULL,
      name TEXT NOT NULL,
      role TEXT DEFAULT 'admin',
      avatar TEXT DEFAULT '',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS courses (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT DEFAULT '',
      instructor TEXT DEFAULT '',
      cover TEXT DEFAULT '',
      category TEXT DEFAULT '',
      status TEXT DEFAULT 'draft',
      student_count INTEGER DEFAULT 0,
      lesson_count INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS students (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      student_no TEXT UNIQUE NOT NULL,
      class_name TEXT DEFAULT '',
      phone TEXT DEFAULT '',
      email TEXT DEFAULT '',
      status TEXT DEFAULT 'active',
      course_ids TEXT DEFAULT '[]',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS learning_records (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      student_id INTEGER,
      course_id INTEGER,
      date TEXT NOT NULL,
      duration INTEGER DEFAULT 0,
      FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
      FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
    );
  `);

  seedData();
}

function seedData() {
  const userCount = db.prepare('SELECT COUNT(*) as count FROM users').get().count;
  if (userCount > 0) return;

  const hashedPassword = bcrypt.hashSync('admin123', 10);
  db.prepare(`
    INSERT INTO users (username, password, name, role) VALUES (?, ?, ?, ?)
  `).run('admin', hashedPassword, '管理员', 'admin');

  const courses = [
    { name: 'React 基础入门', description: '从零开始学习 React 框架，掌握组件化开发思想', instructor: '张老师', category: '前端开发', status: 'published', student_count: 32, lesson_count: 12 },
    { name: 'Node.js 服务端开发', description: '学习 Node.js 构建高性能服务端应用', instructor: '李老师', category: '后端开发', status: 'published', student_count: 28, lesson_count: 10 },
    { name: 'Vue 3 实战项目', description: '通过实际项目掌握 Vue 3 Composition API', instructor: '王老师', category: '前端开发', status: 'published', student_count: 45, lesson_count: 15 },
    { name: 'TypeScript 高级编程', description: '深入理解 TypeScript 类型系统与高级特性', instructor: '赵老师', category: '前端开发', status: 'published', student_count: 20, lesson_count: 8 },
    { name: 'MySQL 数据库设计', description: '数据库设计规范与 SQL 优化实践', instructor: '孙老师', category: '数据库', status: 'published', student_count: 18, lesson_count: 9 },
    { name: 'Docker 容器化部署', description: '学习 Docker 容器技术与微服务部署', instructor: '周老师', category: '运维', status: 'draft', student_count: 0, lesson_count: 6 },
    { name: 'Python 数据分析', description: '使用 Python 进行数据清洗、分析与可视化', instructor: '吴老师', category: '数据科学', status: 'published', student_count: 35, lesson_count: 11 },
    { name: 'Git 版本控制', description: '掌握 Git 工作流与团队协作开发', instructor: '郑老师', category: '工具', status: 'published', student_count: 50, lesson_count: 7 },
    { name: 'Webpack 工程化实践', description: '深入学习 Webpack 配置与前端工程化体系', instructor: '张老师', category: '前端开发', status: 'published', student_count: 22, lesson_count: 9 },
    { name: 'Redis 缓存技术', description: '掌握 Redis 数据结构、持久化与分布式缓存方案', instructor: '李老师', category: '数据库', status: 'published', student_count: 15, lesson_count: 8 },
    { name: 'Linux 运维基础', description: '学习 Linux 常用命令、Shell 脚本与服务器管理', instructor: '周老师', category: '运维', status: 'draft', student_count: 0, lesson_count: 10 },
    { name: 'Jest 单元测试', description: '前端自动化测试框架 Jest 与 React Testing Library 实战', instructor: '赵老师', category: '前端开发', status: 'published', student_count: 12, lesson_count: 6 },
    { name: 'MongoDB 入门到实战', description: '学习 NoSQL 数据库 MongoDB 的 CRUD 与聚合操作', instructor: '孙老师', category: '数据库', status: 'published', student_count: 25, lesson_count: 10 },
  ];

  const insertCourse = db.prepare(`
    INSERT INTO courses (name, description, instructor, category, status, student_count, lesson_count)
    VALUES (@name, @description, @instructor, @category, @status, @student_count, @lesson_count)
  `);

  for (const course of courses) {
    insertCourse.run(course);
  }

  const classNames = ['前端2401班', '前端2402班', '后端2401班', '全栈2401班'];
  const studentNames = [
    '陈明远', '林小雨', '张伟杰', '刘思琪', '王大力',
    '赵文静', '孙志强', '周小红', '吴建国', '郑美玲',
    '黄志勇', '许晓峰', '何雨萱', '胡正阳', '高明月',
    '马思远', '罗晓东', '梁静怡', '谢建华', '宋雅琴',
  ];

  const insertStudent = db.prepare(`
    INSERT INTO students (name, student_no, class_name, phone, email, status, course_ids)
    VALUES (@name, @student_no, @class_name, @phone, @email, @status, @course_ids)
  `);

  for (let i = 0; i < studentNames.length; i++) {
    const courseCount = Math.floor(Math.random() * 3) + 1;
    const courseIds = [];
    while (courseIds.length < courseCount) {
      const id = Math.floor(Math.random() * 12) + 1;
      if (!courseIds.includes(id)) courseIds.push(id);
    }

    insertStudent.run({
      name: studentNames[i],
      student_no: `2024${String(i + 1).padStart(4, '0')}`,
      class_name: classNames[i % classNames.length],
      phone: `138${String(Math.floor(Math.random() * 100000000)).padStart(8, '0')}`,
      email: `student${i + 1}@example.com`,
      status: i < 18 ? 'active' : 'inactive',
      course_ids: JSON.stringify(courseIds),
    });
  }

  const insertRecord = db.prepare(`
    INSERT INTO learning_records (student_id, course_id, date, duration)
    VALUES (@student_id, @course_id, @date, @duration)
  `);

  const today = new Date();
  for (let dayOffset = 6; dayOffset >= 0; dayOffset--) {
    const date = new Date(today);
    date.setDate(date.getDate() - dayOffset);
    const dateStr = date.toISOString().split('T')[0];

    const recordCount = Math.floor(Math.random() * 10) + 5;
    for (let j = 0; j < recordCount; j++) {
      insertRecord.run({
        student_id: Math.floor(Math.random() * 20) + 1,
        course_id: Math.floor(Math.random() * 12) + 1,
        date: dateStr,
        duration: Math.floor(Math.random() * 90) + 10,
      });
    }
  }

  console.log('Mock 数据初始化完成');
}
