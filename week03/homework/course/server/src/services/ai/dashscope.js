const DEFAULT_BASE_URL = 'https://dashscope.aliyuncs.com/compatible-mode/v1';
const DEFAULT_MODEL = 'qwen-plus';

function getApiConfig() {
  const apiKey = process.env.DASHSCOPE_API_KEY || process.env.ALIYUN_BAILIAN_API_KEY || 'sk-e4dcae9060904558a2f64d6eb12249e0';
  const baseUrl = process.env.DASHSCOPE_BASE_URL || DEFAULT_BASE_URL||'https://dashscope.aliyuncs.com/compatible-mode/v1';
  const model = process.env.DASHSCOPE_MODEL || DEFAULT_MODEL||'qwen3.5-plus';
  return { apiKey, baseUrl, model };
}

export function hasDashScopeConfig() {
  const { apiKey } = getApiConfig();
  return Boolean(apiKey);
}

export async function createChatCompletion(payload) {
  const { apiKey, baseUrl, model } = getApiConfig();
  if (!apiKey) {
    throw new Error('未配置 DASHSCOPE_API_KEY（或 ALIYUN_BAILIAN_API_KEY）');
  }

  const response = await fetch(`${baseUrl}/chat/completions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify({
      model,
      temperature: 0.2,
      ...payload,
    }),
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    const msg = data?.error?.message || data?.message || `百炼接口请求失败: ${response.status}`;
    throw new Error(msg);
  }

  return data;
}

/**
 * 流式调用大模型
 * @param {object} payload - 请求参数
 * @param {function} onChunk - 每个chunk回调函数，接收增量文本
 * @returns {Promise<{fullContent: string, usage: object}>}
 */
export async function createStreamingChatCompletion(payload, onChunk) {
  const { apiKey, baseUrl, model } = getApiConfig();
  if (!apiKey) {
    throw new Error('未配置 DASHSCOPE_API_KEY（或 ALIYUN_BAILIAN_API_KEY）');
  }

  const response = await fetch(`${baseUrl}/chat/completions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify({
      model,
      temperature: 0.2,
      stream: true,
      stream_options: { include_usage: true },
      ...payload,
    }),
  });

  if (!response.ok) {
    const text = await response.text();
    let msg = `百炼接口请求失败: ${response.status}`;
    try {
      const data = JSON.parse(text);
      msg = data?.error?.message || data?.message || msg;
    } catch {}
    throw new Error(msg);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder('utf-8');
  let fullContent = '';
  let usage = null;
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });

    // 解析 SSE 格式：每行是 "data: {...}" 或 "data: [DONE]"
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || !trimmed.startsWith('data:')) continue;

      const dataStr = trimmed.slice(5).trim();
      if (dataStr === '[DONE]') continue;

      try {
        const chunk = JSON.parse(dataStr);
        const delta = chunk?.choices?.[0]?.delta;
        if (delta?.content) {
          const increment = delta.content;
          fullContent += increment;
          if (onChunk) {
            onChunk(increment, fullContent);
          }
        }
        if (chunk?.usage) {
          usage = chunk.usage;
        }
      } catch {
        // 忽略解析失败的行
      }
    }
  }

  return { fullContent, usage };
}
