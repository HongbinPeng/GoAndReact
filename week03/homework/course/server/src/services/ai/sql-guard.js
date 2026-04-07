import db from '../../database/db.js';

const WHITELIST_TABLES = new Set(['courses', 'students', 'learning_records']);
const DANGEROUS_KEYWORDS = /\b(insert|update|delete|drop|alter|create|truncate|replace|attach|detach|pragma|vacuum|reindex|grant|revoke)\b/i;
// SQLite 内置表值函数，不应视为表名
const BUILTIN_TABLE_FUNCTIONS = new Set(['json_each', 'json_tree', 'generate_series', 'fts5', 'fts3', 'fts4']);

function normalizeSql(sql) {
  return String(sql || '').trim().replace(/\s+/g, ' ');
}

function extractTables(sql) {
  const tableNames = [];
  const regex = /\b(?:from|join)\s+([`"'[\]\w.]+)/gi;
  let match = regex.exec(sql);
  while (match) {
    const raw = match[1];
    const cleaned = raw
      .replace(/[`"'[\]]/g, '')
      .split('.')
      .pop()
      .toLowerCase();
    // 跳过 SQLite 内置表值函数
    if (cleaned && !BUILTIN_TABLE_FUNCTIONS.has(cleaned)) {
      tableNames.push(cleaned);
    }
    match = regex.exec(sql);
  }
  return [...new Set(tableNames)];
}

export function validateSelectSql(rawSql) {
  const sql = normalizeSql(rawSql);
  if (!sql) {
    return { valid: false, reason: 'SQL 不能为空' };
  }
  if (!/^select\b/i.test(sql)) {
    return { valid: false, reason: '仅允许执行 SELECT 语句' };
  }
  if (sql.includes(';')) {
    return { valid: false, reason: '不允许多语句或分号' };
  }
  if (DANGEROUS_KEYWORDS.test(sql)) {
    return { valid: false, reason: '检测到危险关键字，已拒绝执行' };
  }

  const tableNames = extractTables(sql);
  if (tableNames.length === 0) {
    return { valid: false, reason: 'SQL 未包含可识别的数据表' };
  }
  const illegalTable = tableNames.find((name) => !WHITELIST_TABLES.has(name));
  if (illegalTable) {
    return { valid: false, reason: `表 ${illegalTable} 不在白名单中` };
  }

  return { valid: true, sql, tableNames };
}

export function executeSafeSelect(rawSql) {
  const validation = validateSelectSql(rawSql);
  if (!validation.valid) {
    return { ok: false, error: validation.reason };
  }

  try {
    const rows = db.prepare(validation.sql).all();
    return {
      ok: true,
      sql: validation.sql,
      rows,
      rowCount: rows.length,
      tables: validation.tableNames,
    };
  } catch (error) {
    return {
      ok: false,
      error: `SQL 执行失败: ${error.message}`,
    };
  }
}
