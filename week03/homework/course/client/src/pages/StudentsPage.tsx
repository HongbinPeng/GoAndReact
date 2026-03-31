import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag } from 'antd'
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
    <Card title="学生管理" extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增学生</Button>}>
      <Space style={{ marginBottom: 16 }} wrap>
        <Input.Search placeholder="姓名/学号" allowClear onSearch={(keyword) => setQuery((q) => ({ ...q, page: 1, keyword }))} />
        <Select placeholder="班级" allowClear style={{ width: 180 }} onChange={(className) => setQuery((q) => ({ ...q, page: 1, className: className || '' }))}
          options={classes.map((c) => ({ value: c, label: c }))} />
        <Select placeholder="状态" allowClear style={{ width: 130 }} onChange={(status) => setQuery((q) => ({ ...q, page: 1, status: status || '' }))}
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
          { title: '联系方式', dataIndex: 'phone' },
          { title: '状态', dataIndex: 'status', render: (v: Student['status']) => <Tag color={v === 'active' ? 'green' : 'default'}>{v === 'active' ? '活跃' : '非活跃'}</Tag> },
          { title: '选课数', render: (_, record) => record.course_ids.length },
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
          <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}><Input /></Form.Item>
          <Form.Item label="学号" name="student_no" rules={[{ required: true, message: '请输入学号' }]}><Input /></Form.Item>
          <Form.Item label="班级" name="class_name"><Input /></Form.Item>
          <Form.Item label="电话" name="phone"><Input /></Form.Item>
          <Form.Item label="邮箱" name="email"><Input /></Form.Item>
          <Form.Item label="状态" name="status"><Select options={[{ value: 'active', label: '活跃' }, { value: 'inactive', label: '非活跃' }]} /></Form.Item>
          <Form.Item label="课程" name="course_ids">
            <Select
              mode="multiple"
              options={courses.map((course) => ({ value: course.id, label: course.name }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
