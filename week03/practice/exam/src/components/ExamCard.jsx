function OptionIcon({ variant }) {
  if (variant === 'check') {
    return (
      <span className="option-icon option-icon--ok" aria-hidden>
        ✓
      </span>
    )
  }
  if (variant === 'cross') {
    return (
      <span className="option-icon option-icon--bad" aria-hidden>
        ✕
      </span>
    )
  }
  return null
}

function getOptionState(questionStatus, selectedIndex, correctIndex, optionIndex) {
  if (questionStatus === 'idle') return 'default'
  if (questionStatus === 'timeout') {
    return optionIndex === correctIndex ? 'correct' : 'default'
  }
  if (questionStatus === 'answered') {
    if (optionIndex === correctIndex) return 'correct'
    if (optionIndex === selectedIndex && selectedIndex !== correctIndex) return 'wrong'
    return 'default'
  }
  return 'default'
}

export default function ExamCard({
  title,
  questionNumber,
  totalQuestions,
  options,
  remainingSeconds,
  questionStatus,
  selectedIndex,
  correctIndex,
  onSelectOption,
  onNext,
}) {
  const canGoNext = questionStatus !== 'idle'

  return (
    <div className="card exam-card">
      <header className="exam-header">
        <div className="exam-header-text">
          <h1 className="exam-app-title">考试小应用</h1>
          <div className="exam-title-underline" />
        </div>
        <div className="timer-block" aria-live="polite">
          <span className="timer-label">剩余时间</span>
          <span className="timer-value">{remainingSeconds}</span>
        </div>
      </header>

      <p className="question-text">
        {questionNumber}. {title}
      </p>

      <ul className="options-list">
        {options.map((label, idx) => {
          const state = getOptionState(questionStatus, selectedIndex, correctIndex, idx)
          const icon = state === 'correct' ? 'check' : state === 'wrong' ? 'cross' : null
          return (
            <li key={idx}>
              <button
                type="button"
                className={`option-btn option-btn--${state}`}
                onClick={() => onSelectOption(idx)}
                disabled={questionStatus !== 'idle'}
              >
                <span className="option-label">{label}</span>
                <OptionIcon variant={icon} />
              </button>
            </li>
          )
        })}
      </ul>

      <footer className="exam-footer">
        <span className="progress-text">
          进度：{questionNumber} / {totalQuestions}
        </span>
        <button type="button" className="btn btn-primary" disabled={!canGoNext} onClick={onNext}>
          下一题
        </button>
      </footer>
    </div>
  )
}
