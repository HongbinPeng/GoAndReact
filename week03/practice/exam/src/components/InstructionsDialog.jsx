export default function InstructionsDialog({ onContinue, onExit }) {
  return (
    <div className="modal-backdrop" role="presentation">
      <div className="modal card instructions-card" role="dialog" aria-modal="true" aria-labelledby="instructions-title">
        <h2 id="instructions-title" className="instructions-title">
          考试须知
        </h2>
        <ul className="instructions-list">
          <li>每道题答题时间为 <strong>15 秒</strong>，请在倒计时结束前选择答案。</li>
          <li>提交答案或超时后，本轮倒计时会停止，请通过「下一题」继续。</li>
          <li>请独立完成考试，勿刷新页面以免进度丢失。</li>
          <li>全部题目作答结束后，将显示得分与评价。</li>
        </ul>
        <div className="modal-actions">
          <button type="button" className="btn btn-outline" onClick={onExit}>
            退出考试
          </button>
          <button type="button" className="btn btn-primary" onClick={onContinue}>
            继续考试
          </button>
        </div>
      </div>
    </div>
  )
}
