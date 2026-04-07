import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import AppLayout from './layouts/AppLayout'
import DashboardPage from './pages/DashboardPage'
import CoursesPage from './pages/CoursesPage'
import StudentsPage from './pages/StudentsPage'
import SummaryPage from './pages/SummaryPage'
import AiChatPage from './pages/AiChatPage'
import LoginPage from './pages/LoginPage'
import { getToken } from './utils/auth'

function ProtectedRoute() {
  if (!getToken()) {
    return <Navigate to="/login" replace />
  }
  return <AppLayout />
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<ProtectedRoute />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="courses" element={<CoursesPage />} />
          <Route path="students" element={<StudentsPage />} />
          <Route path="summary" element={<SummaryPage />} />
          <Route path="ai-chat" element={<AiChatPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
