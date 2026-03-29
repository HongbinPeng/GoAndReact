export default function StartCard({ onStart }) {
  return (
    <div className="card start-card">
      <button type="button" className="btn-text btn-start" onClick={onStart}>
        开始考试
      </button>
    </div>
  )
}
