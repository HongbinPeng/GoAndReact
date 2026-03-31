import http from './http'
import type { Course, DashboardData, PaginationData, Student } from '../types'

export const api = {
  login: (payload: { username: string; password: string }) =>
    http.post<{ token: string; user: { id: number; username: string; name: string } }>('/auth/login', payload),

  getDashboard: () => http.get<DashboardData>('/dashboard'),

  getCourseList: (params: Record<string, unknown>) =>
    http.get<PaginationData<Course>>('/courses', { params }),
  getCourseCategories: () => http.get<string[]>('/courses/categories'),
  createCourse: (payload: Partial<Course>) => http.post<Course>('/courses', payload),
  updateCourse: (id: number, payload: Partial<Course>) => http.put<Course>(`/courses/${id}`, payload),
  deleteCourse: (id: number) => http.delete(`/courses/${id}`),
  toggleCourseStatus: (id: number) => http.patch<Course>(`/courses/${id}/status`),

  getStudentList: (params: Record<string, unknown>) =>
    http.get<PaginationData<Student>>('/students', { params }),
  getClasses: () => http.get<string[]>('/students/classes'),
  createStudent: (payload: Partial<Student>) => http.post<Student>('/students', payload),
  updateStudent: (id: number, payload: Partial<Student>) => http.put<Student>(`/students/${id}`, payload),
  deleteStudent: (id: number) => http.delete(`/students/${id}`),

  getSummary: () => http.get<{ content: string }>('/summary'),
}
