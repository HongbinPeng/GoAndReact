import { useCallback, useEffect, useState } from 'react'
import { questions } from './data/questions'
import StartCard from './components/StartCard'
import InstructionsDialog from './components/InstructionsDialog'
import ExamCard from './components/ExamCard'
import ResultCard from './components/ResultCard'

const TOTAL = questions.length

export default function App() {
  const [view, setView] = useState('start')
  const [instructionsOpen, setInstructionsOpen] = useState(false)

  const [currentIndex, setCurrentIndex] = useState(0)
  const [remainingSeconds, setRemainingSeconds] = useState(15)
  const [questionStatus, setQuestionStatus] = useState('idle')
  const [selectedIndex, setSelectedIndex] = useState(null)

  const [correctCount, setCorrectCount] = useState(0)

  const ratePercent = TOTAL > 0 ? Math.round((correctCount / TOTAL) * 100) : 0

  const resetExamSession = useCallback(() => {
    setCurrentIndex(0)
    setRemainingSeconds(15)
    setQuestionStatus('idle')
    setSelectedIndex(null)
    setCorrectCount(0)
  }, [])

  const goHome = useCallback(() => {
    setView('start')
    setInstructionsOpen(false)
    resetExamSession()
  }, [resetExamSession])

  useEffect(() => {
    if (view !== 'exam' || questionStatus !== 'idle') return
    const id = setInterval(() => {
      setRemainingSeconds((s) => (s <= 1 ? 0 : s - 1))
    }, 1000)
    return () => clearInterval(id)
  }, [view, questionStatus, currentIndex])

  useEffect(() => {
    if (view !== 'exam') return
    if (questionStatus !== 'idle') return
    if (remainingSeconds !== 0) return
    setQuestionStatus('timeout')
  }, [remainingSeconds, view, questionStatus])

  const handleStartClick = () => {
    setInstructionsOpen(true)
  }

  const handleExitInstructions = () => {
    setInstructionsOpen(false)
  }

  const handleContinueExam = () => {
    setInstructionsOpen(false)
    resetExamSession()
    setView('exam')
  }

  const handleSelectOption = (idx) => {
    if (view !== 'exam' || questionStatus !== 'idle') return
    const q = questions[currentIndex]
    setSelectedIndex(idx)
    setQuestionStatus('answered')
    if (idx === q.correctIndex) setCorrectCount((c) => c + 1)
  }

  const handleNext = () => {
    if (questionStatus === 'idle') return
    if (currentIndex >= TOTAL - 1) {
      setView('result')
      return
    }
    setCurrentIndex((i) => i + 1)
    setRemainingSeconds(15)
    setQuestionStatus('idle')
    setSelectedIndex(null)
  }

  const currentQuestion = questions[currentIndex]

  return (
    <div className="app-shell">
      {view === 'start' && !instructionsOpen && <StartCard onStart={handleStartClick} />}
      {instructionsOpen && (
        <InstructionsDialog onContinue={handleContinueExam} onExit={handleExitInstructions} />
      )}
      {view === 'exam' && currentQuestion && (
        <ExamCard
          title={currentQuestion.title}
          questionNumber={currentIndex + 1}
          totalQuestions={TOTAL}
          options={currentQuestion.options}
          remainingSeconds={remainingSeconds}
          questionStatus={questionStatus}
          selectedIndex={selectedIndex}
          correctIndex={currentQuestion.correctIndex}
          onSelectOption={handleSelectOption}
          onNext={handleNext}
        />
      )}
      {view === 'result' && (
        <ResultCard
          correctCount={correctCount}
          totalQuestions={TOTAL}
          ratePercent={ratePercent}
          onRetry={goHome}
          onExit={goHome}
        />
      )}
    </div>
  )
}
