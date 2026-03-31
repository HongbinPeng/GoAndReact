import { Card, Col, Empty, Row, Statistic } from 'antd'
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
    <>
      <Row gutter={16}>
        <Col span={6}><Card><Statistic title="课程总数" value={data.stats.totalCourses} /></Card></Col>
        <Col span={6}><Card><Statistic title="学生总数" value={data.stats.totalStudents} /></Card></Col>
        <Col span={6}><Card><Statistic title="发布率" value={publishRate} suffix="%" /></Card></Col>
        <Col span={6}><Card><Statistic title="活跃率" value={activeRate} suffix="%" /></Card></Col>
      </Row>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="课程选课人数排行">
            {data.charts.enrollment.length ? (
              <ReactECharts option={{
                xAxis: { type: 'category', data: data.charts.enrollment.map((item) => item.name) },
                yAxis: { type: 'value' },
                tooltip: {},
                series: [{ type: 'bar', data: data.charts.enrollment.map((item) => item.value) }],
              }} />
            ) : <Empty />}
          </Card>
        </Col>
        <Col span={12}>
          <Card title="近7天学习活跃度">
            <ReactECharts option={{
              tooltip: { trigger: 'axis' },
              legend: { data: ['学习人数', '学习时长(小时)'] },
              xAxis: { type: 'category', data: data.charts.activity.map((item) => item.label) },
              yAxis: { type: 'value' },
              series: [
                { name: '学习人数', type: 'line', data: data.charts.activity.map((item) => item.students) },
                { name: '学习时长(小时)', type: 'line', data: data.charts.activity.map((item) => item.duration) },
              ],
            }} />
          </Card>
        </Col>
      </Row>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="学生状态分布">
            <ReactECharts option={{
              tooltip: { trigger: 'item' },
              series: [{ type: 'pie', radius: '60%', data: data.charts.statusDist }],
            }} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="课程分类分布">
            <ReactECharts option={{
              tooltip: { trigger: 'item' },
              series: [{ type: 'pie', radius: '60%', data: data.charts.categoryDist }],
            }} />
          </Card>
        </Col>
      </Row>
    </>
  )
}
