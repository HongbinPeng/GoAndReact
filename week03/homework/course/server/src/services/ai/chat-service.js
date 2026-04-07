import { createChatCompletion, createStreamingChatCompletion } from './dashscope.js';
import { queryAllCourses, queryAllLearningRecords, queryAllStudents } from './db-tools.js';
import { executeSafeSelect } from './sql-guard.js';

/**
 * 获取三张表的字段信息和示例数据（各取第一行）
 */
function getSchemaExamples() {
  const courses = queryAllCourses();
  const students = queryAllStudents();
  const learningRecords = queryAllLearningRecords();

  return {
    courses: {
      fields: courses.rows.length > 0 ? Object.keys(courses.rows[0]) : [],
      example: courses.rows[0] || null,
      count: courses.count,
    },
    students: {
      fields: students.rows.length > 0 ? Object.keys(students.rows[0]) : [],
      example: students.rows[0] || null,
      count: students.count,
    },
    learningRecords: {
      fields: learningRecords.rows.length > 0 ? Object.keys(learningRecords.rows[0]) : [],
      example: learningRecords.rows[0] || null,
      count: learningRecords.count,
    },
  };
}

/**
 * 使用大模型判断是否需要查询数据库并生成 SQL
 * 返回：{ needQuery: boolean, sql?: string }
 */
async function analyzeQueryIntent(question, schemaExamples) {
  const prompt = [
    '你是课程管理系统的 SQL 生成助手。',
    '请根据用户问题和数据库表结构，判断是否需要查询数据库来回答。',
    '',
    '如果不需要查询数据库（闲聊、问候、主观评价等），返回：{"needQuery": false}',
    '如果需要查询数据库，返回：{"needQuery": true, "sql": "SELECT ..."}',
    '',
    '约束：',
    '- 只能生成一条 SELECT 语句',
    '- 禁止分号',
    '- 仅能访问 courses、students、learning_records 三张表',
    '- SQL 必须符合 SQLite 语法',
    '',
    '重要说明 - students.course_ids 字段：',
    '- 字段类型：TEXT，存储 JSON 数组格式的字符串，例如 "[1,2,3]"',
    '- 字段含义：存储学生已选修的所有课程 ID 列表',
    '- 查询方法 1（不推荐，这样会造成歧义，例如查询课程 ID=1 的学生，会模糊匹配到课程id为11的课程）：WHERE course_ids LIKE "%[课程 ID]%" 或 WHERE course_ids LIKE "%课程 ID%"',
    '- 查询方法 2（标准 JSON，推荐）：使用 JSON_EACH(course_ids) 展开，value 即为课程 ID',
    '- 示例：查询选修课程 ID=1 的学生（不推荐） -> SELECT * FROM students WHERE course_ids LIKE "%1%"',
    '- 示例：统计每门课程人数 -> SELECT c.id, c.name, COUNT(*) FROM courses c, students s, JSON_EACH(s.course_ids) WHERE CAST(value AS INTEGER) = c.id GROUP BY c.id',
    '-注意：learning_records 表仅代表学生学习记录，不包含选课关系，真正的选课关系在 students 表的 course_ids 字段',
    '数据库表结构：',
    `（1）courses 课程主表
字段名，字段类型，字段含义
id	INTEGER	课程唯一主键 ID，自增整数，全局唯一标识一门课程，用于和其他表做关联
name	TEXT	课程全称，用于前端展示课程名称，如「React 基础入门」
description	TEXT	课程简介/课程描述，说明课程的学习目标、核心内容、适用人群等
instructor	TEXT	课程授课讲师姓名，如「张老师」「李老师」
category	TEXT	课程分类标签，用于课程的归类与筛选，如「前端开发」「后端开发」「数据库」「运维」等
status	TEXT	课程发布状态，枚举值：published（已发布，对学生可见）、draft（草稿，仅后台可见）
student_count	INTEGER	该课程的学习/报名学生总人数，用于统计课程热度
lesson_count	INTEGER	该课程包含的总课时数，标识课程的内容体量
created_at	DATETIME	课程记录的创建时间，即课程信息录入系统的时间
updated_at	DATETIME	课程信息的最后更新时间，课程内容、状态修改时同步更新`,
    `（2）learning_records 学生学习记录表，业务定位：记录学生的课程学习行为数据，是学生与课程的关联中间表，用于统计学习时长、学习轨迹。
字段名	字段类型	字段含义
id	INTEGER	学习记录唯一主键 ID，自增整数，全局唯一标识一条学习行为记录
student_id	INTEGER	学生 ID，关联 students 表的 id 字段，标识该条记录所属的学生
course_id	INTEGER	课程 ID，关联 courses 表的 id 字段，标识学生本次学习的课程
date	TEXT	本次学习的发生日期，格式为 YYYY-MM-DD，用于按日期统计学习行为
duration	INTEGER	本次学习的有效时长，单位为分钟，用于统计学生的学习投入情况`,
    `（3）students 学生信息表：存储平台所有学生的个人基础信息、账号状态、选课情况，是系统的用户核心表。
字段名	字段类型	字段含义
id	INTEGER	学生唯一主键 ID，自增整数，全局唯一标识一名学生，与其他表的 student_id 关联
name	TEXT	学生的真实姓名
student_no	TEXT	学生学号，系统内唯一的学籍编号，如 20240001，用于学籍管理、身份核验
class_name	TEXT	学生所属的班级名称，如「前端 2401 班」「全栈 2401 班」，用于班级维度的学生管理
phone	TEXT	学生的联系手机号，用于账号绑定、通知触达
email	TEXT	学生的邮箱地址，用于账号登录、密码找回、系统通知
status	TEXT	学生状态，枚举值：active（学生比较活跃）、inactive（学生未活跃）
course_ids	TEXT	学生已报名/已选的课程 ID 列表，数组格式存储，关联 courses 表的 id 字段，记录学生的选课情况
created_at	DATETIME	学生账号/信息的创建时间，即学生信息录入系统、账号注册的时间
updated_at	DATETIME	学生信息的最后更新时间，个人信息、选课情况、账号状态修改时同步更新`,
    '',
    '某条实际数据示例：',
    `- courses 表 (${schemaExamples.courses.count}条数据): 字段 ${schemaExamples.courses.fields.join(', ')}, 示例：${JSON.stringify(schemaExamples.courses.example)}`,
    `- students 表 (${schemaExamples.students.count}条数据): 字段 ${schemaExamples.students.fields.join(', ')}, 示例：${JSON.stringify(schemaExamples.students.example)}`,
    `- learning_records 表 (${schemaExamples.learningRecords.count}条数据): 字段 ${schemaExamples.learningRecords.fields.join(', ')}, 示例：${JSON.stringify(schemaExamples.learningRecords.example)}`,
    '',
    `用户问题：${question}`,
  ].join('\n');

  const completion = await createChatCompletion({
    messages: [{ role: 'user', content: prompt }],
    temperature: 0,
  });

  const text = completion?.choices?.[0]?.message?.content || '{}';

  // 尝试提取 JSON
  try {
    const jsonMatch = text.match(/\{[\s\S]*\}/);
    if (jsonMatch) {
      return JSON.parse(jsonMatch[0]);
    }
  } catch {}

  // 如果解析失败，检查是否包含 SQL 关键字
  const upperText = text.toUpperCase();
  if (upperText.includes('SELECT')) {
    return { needQuery: true, sql: text };
  }

  return { needQuery: false };
}

function parseSqlFromText(text) {
  const content = String(text || '').trim();
  const fencedMatch = content.match(/```sql\s*([\s\S]*?)```/i);
  if (fencedMatch) {
    return fencedMatch[1].trim();
  }
  const jsonMatch = content.match(/\{[\s\S]*"sql"\s*:\s*"([\s\S]*?)"[\s\S]*\}/i);
  if (jsonMatch) {
    return jsonMatch[1].replace(/\\"/g, '"').trim();
  }
  return content;
}

function pushTrace(trace, step, detail, sql) {
  trace.push({
    step,
    detail,
    sql: sql || null,
    at: new Date().toISOString(),
  });
}

/**
 * 流式 AI 对话
 * @param {object} params - { question, sessionId }
 * @param {function} onChunk - 流式回调，接收 (increment, fullContent)
 * @returns {Promise<{trace: array, sqlInfo?: object}>}
 */
export async function chatWithAiStreaming({ question, sessionId }, onChunk) {
  const trace = [];
  pushTrace(trace, 'request_received', `sessionId=${sessionId || 'new'}`);

  // 第一步：获取三张表的字段信息和示例数据
  const schemaExamples = getSchemaExamples();
  pushTrace(trace, 'schema_fetched', `courses:${schemaExamples.courses.count} students:${schemaExamples.students.count} learning_records:${schemaExamples.learningRecords.count}`);

  // 第二步：让大模型判断是否需要查询数据库并生成 SQL
  const analyzeResult = await analyzeQueryIntent(question, schemaExamples);
  pushTrace(trace, 'intent_analyzed', `needQuery=${analyzeResult.needQuery}${analyzeResult.sql ? ', sql_generated' : ''}`, analyzeResult.sql || null);

  // 第三步：不需要查询数据库，直接流式回复
  if (!analyzeResult.needQuery) {
    const prompt = [
      { role: 'system', content: '你是课程管理系统的中文 AI 助手，请用 markdown 格式回答用户问题，回答要简洁清晰。' },
      { role: 'user', content: question },
    ];

    await createStreamingChatCompletion({ messages: prompt }, onChunk);
    pushTrace(trace, 'streaming_chat', '流式问答已完成');
    return { trace };
  }

  // 第四步：执行 SQL 查询
  const sql = parseSqlFromText(analyzeResult.sql || '');
  const execution = executeSafeSelect(sql);
  if (!execution.ok) {
    pushTrace(trace, 'sql_blocked', execution.error);
    throw new Error(execution.error);
  }
  pushTrace(trace, 'sql_executed', `命中表 ${execution.tables.join(', ')}，返回 ${execution.rowCount} 条`);

  // 第五步：流式生成最终回答
  const prompt = [
    '你是课程管理系统助手，请根据查询结果给出用户可读的回答。',
    '输出要求：',
    '- 包含查询结论、关键数据、总条数',
    '- 使用 markdown 格式，可以用表格展示数据',
    '- 回答要简洁清晰',
    '',
    `用户问题：${question}`,
    `执行 SQL：${execution.sql}`,
    `查询结果（共${execution.rows.length}条）：${JSON.stringify(execution.rows.slice(0, 50))}`, // 限制数据量
  ].join('\n');

  await createStreamingChatCompletion({ messages: [{ role: 'user', content: prompt }] }, onChunk);
  pushTrace(trace, 'streaming_answer', '流式回答已生成');

  return {
    trace,
    sqlInfo: {
      sql: execution.sql,
      rowCount: execution.rowCount,
      tables: execution.tables,
    },
  };
}