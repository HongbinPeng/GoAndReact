import completedExamImage from '../assets/完成考试.png'

function CrownIcon() {
  return (
    <img 
      src={completedExamImage} 
      alt="完成考试" 
      className="result-crown-svg"
      style={{ width: '72px', height: '54px' }}
    />
  )
}

function tierHint(ratePercent, isPassed) {
  if (isPassed) return ''
  if (ratePercent <= 20) return '不要灰心，多花点时间巩固基础，下次一定更好。'
  if (ratePercent <= 60) return '很好，下次继续加油！'
  return ''
}

/** 与需求一致：正确率 strictly > 60% 视为通过 */
export default function ResultCard({
  correctCount,
  totalQuestions,
  ratePercent,
  onRetry,
  onExit,
}) {
  const isPassed = ratePercent > 60
  const hint = tierHint(ratePercent, isPassed)

  return (
    <div className="card result-card result-card--unified">
      <div className="result-crown-wrap">
        <CrownIcon />
      </div>
      <h2 className="result-done-title">您已完成考试！</h2>
      <p className="result-score-line-unified">
        {isPassed ? (
          <>
            恭喜！🎉，你得了 <strong>{correctCount}</strong> 分，共 <strong>{totalQuestions}</strong> 分。
          </>
        ) : (
          <>
            你得了 <strong>{correctCount}</strong> 分，共 <strong>{totalQuestions}</strong> 分。
          </>
        )}
      </p>
      <p className="result-rate-unified">
        正确率：<span className="result-rate-value">{ratePercent}%</span>
        {isPassed ? <span className="result-pass-tag">已通过</span> : <span className="result-fail-tag">未通过</span>}
      </p>
      {hint ? <p className="result-hint">{hint}</p> : null}
      <div className="result-actions result-actions--unified">
        <button type="button" className="btn btn-primary btn-result-primary" onClick={onRetry}>
          重新考试
        </button>
        <button type="button" className="btn btn-outline btn-result-outline" onClick={onExit}>
          退出考试
        </button>
      </div>
    </div>
  )
}
