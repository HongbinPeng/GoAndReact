import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Checkbox, Col, Form, Input, Modal, Popconfirm, Row, Select, Space, Table, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Course, Student } from '../types'

type FormValues = Partial<Student>

export default function StudentsPage() {
  const [list, setList] = useState<Student[]>([])
  const [classes, setClasses] = useState<string[]>([])
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [query, setQuery] = useState({ page: 1, pageSize: 10, keyword: '', status: '', className: '' })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Student | null>(null)
  const [form] = Form.useForm<FormValues>()

  const load = async () => {
    setLoading(true)
    try {
      const data = await api.getStudentList(query)
      setList(data.list)
      setTotal(data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [query])

  useEffect(() => {
    api.getClasses().then(setClasses)
    api.getCourseList({ page: 1, pageSize: 200 }).then((res) => setCourses(res.list))
  }, [])

  return (
    <div className="page-wrap">
      <div className="page-head">
        <Typography.Title level={4}>学生管理</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增学生</Button>
      </div>
      <Card className="panel-card">
      <Space className="filter-bar" wrap>
        <Input.Search placeholder="请输入姓名/学号" allowClear onSearch={(keyword) => setQuery((q) => ({ ...q, page: 1, keyword }))} />
        <Select placeholder="班级筛选" allowClear style={{ width: 180 }} onChange={(className) => setQuery((q) => ({ ...q, page: 1, className: className || '' }))}
          options={classes.map((c) => ({ value: c, label: c }))} />
        <Select placeholder="学生状态" allowClear style={{ width: 130 }} onChange={(status) => setQuery((q) => ({ ...q, page: 1, status: status || '' }))}
          options={[{ value: 'active', label: '活跃' }, { value: 'inactive', label: '非活跃' }]} />
      </Space>
      <Table<Student>
        rowKey="id"
        loading={loading}
        dataSource={list}
        onChange={(pagination) => setQuery((q) => ({ ...q, page: pagination.current || 1, pageSize: pagination.pageSize || 10 }))}
        pagination={{ current: query.page, pageSize: query.pageSize, total, showSizeChanger: true }}
        columns={[
          { title: '姓名', dataIndex: 'name' },
          { title: '学号', dataIndex: 'student_no' },
          { title: '班级', dataIndex: 'class_name' },
          { title: '联系方式', render: (_, record) => <div>{record.phone}<br />{record.email}</div> },
          { title: '状态', dataIndex: 'status', render: (v: Student['status']) => <Tag color={v === 'active' ? 'green' : 'default'}>{v === 'active' ? '活跃' : '非活跃'}</Tag> },
          {
            title: '已选课程',
            render: (_, record) => {
              const text = courses.filter((c) => record.course_ids.includes(c.id)).map((c) => c.name).join('、')
              return <span>{text || '-'}</span>
            },
          },
          {
            title: '操作',
            render: (_, record) => (
              <Space>
                <Button
                  icon={<EditOutlined />}
                  onClick={() => {
                    setEditing(record)
                    form.setFieldsValue(record)
                    setOpen(true)
                  }}
                >
                  编辑
                </Button>
                <Popconfirm title="确认删除该学生？" onConfirm={async () => { await api.deleteStudent(record.id); await load() }}>
                  <Button danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      className="data-table"
      />

      <Modal
        title={editing ? '编辑学生' : '新增学生'}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={async () => {
          const values = await form.validateFields()
          if (editing) {
            await api.updateStudent(editing.id, values)
          } else {
            await api.createStudent(values)
          }
          setOpen(false)
          await load()
        }}
      >
        <Form form={form} layout="vertical" initialValues={{ status: 'active', course_ids: [] }}>
          <Row gutter={12}>
            <Col span={12}><Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}><Input /></Form.Item></Col>
            <Col span={12}><Form.Item label="学号" name="student_no" rules={[{ required: true, message: '请输入学号' }]}><Input /></Form.Item></Col>
            <Col span={12}><Form.Item label="班级" name="class_name"><Input /></Form.Item></Col>
            <Col span={12}><Form.Item label="状态" name="status"><Select options={[{ value: 'active', label: '活跃' }, { value: 'inactive', label: '非活跃' }]} /></Form.Item></Col>
            <Col span={12}><Form.Item label="手机号" name="phone"><Input /></Form.Item></Col>
            <Col span={12}><Form.Item label="邮箱" name="email"><Input /></Form.Item></Col>
          </Row>
          <Form.Item label="课程" name="course_ids">
            <Checkbox.Group className="course-check-grid" options={courses.map((course) => ({ value: course.id, label: course.name }))} />
          </Form.Item>
        </Form>
      </Modal>
      </Card>
    </div>
  )
}
