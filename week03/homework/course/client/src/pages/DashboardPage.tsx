import { Card, Col, Empty, Row, Typography } from 'antd'
import ReactECharts from 'echarts-for-react'
import { useEffect, useState } from 'react'
import { api } from '../api'
import type { DashboardData } from '../types'

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null)

  useEffect(() => {
    api.getDashboard().then(setData)
  }, [])

  if (!data) return <Card loading />

  const publishRate = data.stats.totalCourses ? Math.round((data.stats.publishedCourses / data.stats.totalCourses) * 100) : 0
  const activeRate = data.stats.totalStudents ? Math.round((data.stats.activeStudents / data.stats.totalStudents) * 100) : 0

  return (
    <div className="page-wrap">
      <div className="page-head">
        <Typography.Title level={4}>工作台</Typography.Title>
      </div>
      <Row gutter={16}>
        <Col span={6}><Card className="stat-card"><div className="stat-name">📊 课程总数</div><div className="stat-value">{data.stats.totalCourses}</div><div className="stat-sub">/ 已发布 {data.stats.publishedCourses}</div></Card></Col>
        <Col span={6}><Card className="stat-card"><div className="stat-name">👥 学生总数</div><div className="stat-value">{data.stats.totalStudents}</div><div className="stat-sub">/ 活跃 {data.stats.activeStudents}</div></Card></Col>
        <Col span={6}><Card className="stat-card"><div className="stat-name">📈 课程发布率</div><div className="stat-value">{publishRate}%</div><div className="stat-sub">&nbsp;</div></Card></Col>
        <Col span={6}><Card className="stat-card"><div className="stat-name">🔥 学生活跃率</div><div className="stat-value">{activeRate}%</div><div className="stat-sub">&nbsp;</div></Card></Col>
      </Row>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="课程选课人数排行" className="panel-card">
            {data.charts.enrollment.length ? (
              <div className="chart-shell"><ReactECharts option={{
                xAxis: { type: 'category', data: data.charts.enrollment.map((item) => item.name) },
                yAxis: { type: 'value' },
                tooltip: {},
                series: [{ type: 'bar', data: data.charts.enrollment.map((item) => item.value) }],
              }} /></div>
            ) : <Empty />}
          </Card>
        </Col>
        <Col span={12}>
          <Card title="近7天学习活跃度" className="panel-card">
            <div className="chart-shell"><ReactECharts option={{
              tooltip: { trigger: 'axis' },
              legend: { data: ['学习人数', '学习时长(小时)'] },
              xAxis: { type: 'category', data: data.charts.activity.map((item) => item.label) },
              yAxis: { type: 'value' },
              series: [
                { name: '学习人数', type: 'line', data: data.charts.activity.map((item) => item.students) },
                { name: '学习时长(小时)', type: 'line', data: data.charts.activity.map((item) => item.duration) },
              ],
            }} /></div>
          </Card>
        </Col>
      </Row>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="学生状态分布" className="panel-card">
            <div className="chart-shell"><ReactECharts option={{
              tooltip: { trigger: 'item' },
              series: [{ type: 'pie', radius: '60%', data: data.charts.statusDist }],
            }} /></div>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="课程分类分布" className="panel-card">
            <div className="chart-shell"><ReactECharts option={{
              tooltip: { trigger: 'item' },
              series: [{ type: 'pie', radius: '60%', data: data.charts.categoryDist }],
            }} /></div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
