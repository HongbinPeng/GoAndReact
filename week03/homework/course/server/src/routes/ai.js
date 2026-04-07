import Router from '@koa/router';
import { authenticateToken } from '../middleware/auth.js';
import { fail } from '../utils/response.js';
import { hasDashScopeConfig } from '../services/ai/dashscope.js';
import { chatWithAiStreaming } from '../services/ai/chat-service.js';

const router = new Router();

/**
 * 流式 AI 对话接口 (SSE)
 */
router.post('/chat/stream', authenticateToken, async (ctx) => {
  const { question = '', sessionId = '' } = ctx.request.body || {};
  const normalizedQuestion = String(question).trim();

  if (!normalizedQuestion) {
    return fail(ctx, 400, 'question 不能为空');
  }
  if (!hasDashScopeConfig()) {
    return fail(ctx, 500, '未配置百炼 API Key，请设置 DASHSCOPE_API_KEY');
  }

  // 设置 SSE 响应头
  ctx.set({
    'Content-Type': 'text/event-stream; charset=utf-8',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'X-Accel-Buffering': 'no', // 禁用 nginx 缓冲
  });

  ctx.status = 200;
  ctx.respond = false; // 让 Koa 不自动处理响应

  const writeSSE = (event, data) => {
    ctx.res.write(`event: ${event}\n`);
    ctx.res.write(`data: ${JSON.stringify(data)}\n\n`);
  };

  try {
    // 发送开始事件
    writeSSE('start', {
      sessionId: sessionId || `${Date.now()}`,
      question: normalizedQuestion,
    });

    // 流式生成回答
    const result = await chatWithAiStreaming(
      { question: normalizedQuestion, sessionId },
      (increment, fullContent) => {
        writeSSE('chunk', { increment, fullContent });
      }
    );

    // 发送完成事件
    writeSSE('done', {
      trace: result.trace,
      sqlInfo: result.sqlInfo || null,
    });
  } catch (error) {
    writeSSE('error', { message: error.message });
  } finally {
    ctx.res.end();
  }
});

export default router;
