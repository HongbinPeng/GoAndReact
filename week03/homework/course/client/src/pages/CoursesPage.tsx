import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { Button, Card, Col, Form, Input, InputNumber, Modal, Popconfirm, Row, Select, Space, Switch, Table, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Course } from '../types'

type FormValues = Partial<Course>

export default function CoursesPage() {
  const [list, setList] = useState<Course[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [query, setQuery] = useState({ page: 1, pageSize: 10, keyword: '', status: '', category: '' })
  const [sorter, setSorter] = useState({ sortField: '', sortOrder: '' })
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Course | null>(null)
  const [form] = Form.useForm<FormValues>()

  const load = async () => {
    setLoading(true)
    try {
      const data = await api.getCourseList({ ...query, ...sorter })
      setList(data.list)
      setTotal(data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [query, sorter])

  useEffect(() => {
    api.getCourseCategories().then(setCategories)
  }, [])

  return (
    <div className="page-wrap">
      <div className="page-head">
        <Typography.Title level={4}>课程管理</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增课程</Button>
      </div>
      <Card className="panel-card">
      <Space className="filter-bar" wrap>
        <Input.Search placeholder="请输入课程名称/讲师" allowClear onSearch={(keyword) => setQuery((q) => ({ ...q, page: 1, keyword }))} />
        <Select placeholder="课程状态" allowClear style={{ width: 140 }} onChange={(status) => setQuery((q) => ({ ...q, page: 1, status: status || '' }))}
          options={[{ value: 'published', label: '已发布' }, { value: 'draft', label: '草稿' }]} />
        <Select placeholder="课程分类" allowClear style={{ width: 160 }} onChange={(category) => setQuery((q) => ({ ...q, page: 1, category: category || '' }))}
          options={categories.map((c) => ({ value: c, label: c }))} />
      </Space>
      <Table<Course>
        rowKey="id"
        loading={loading}
        dataSource={list}
        onChange={(pagination, _filters, sort) => {
          setQuery((q) => ({ ...q, page: pagination.current || 1, pageSize: pagination.pageSize || 10 }))
          if (!Array.isArray(sort)) {
            setSorter({ sortField: (sort.field as string) || '', sortOrder: (sort.order as string) || '' })
          }
        }}
        pagination={{ current: query.page, pageSize: query.pageSize, total, showSizeChanger: true }}
        columns={[
          { title: '课程名称', dataIndex: 'name' },
          { title: '讲师', dataIndex: 'instructor' },
          { title: '分类', dataIndex: 'category' },
          { title: '课时', dataIndex: 'lesson_count' },
          { title: '选课人数', dataIndex: 'student_count', sorter: true },
          { title: '状态', dataIndex: 'status', render: (v: Course['status']) => <Tag color={v === 'published' ? 'green' : 'default'}>{v === 'published' ? '已发布' : '草稿'}</Tag> },
          {
            title: '发布切换',
            render: (_, record) => (
              <Switch checked={record.status === 'published'} onChange={async () => { await api.toggleCourseStatus(record.id); await load() }} />
            ),
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
                <Popconfirm title="确认删除该课程？" onConfirm={async () => { await api.deleteCourse(record.id); await load() }}>
                  <Button danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      className="data-table"
      />

      <Modal
        title={editing ? '编辑课程' : '新增课程'}
        open={open}
        onCancel={() => setOpen(false)}
        onOk={async () => {
          const values = await form.validateFields()
          if (editing) {
            await api.updateCourse(editing.id, values)
          } else {
            await api.createCourse(values)
          }
          setOpen(false)
          await load()
        }}
      >
        <Form form={form} layout="vertical" initialValues={{ status: 'draft', lesson_count: 0 }}>
          <Form.Item label="课程名称" name="name" rules={[{ required: true, message: '请输入课程名称' }]}><Input /></Form.Item>
          <Form.Item label="讲师" name="instructor"><Input /></Form.Item>
          <Form.Item label="分类" name="category"><Input /></Form.Item>
          <Row gutter={12}>
            <Col span={12}><Form.Item label="课时数" name="lesson_count"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={12}><Form.Item label="状态" name="status"><Select options={[{ value: 'draft', label: '草稿' }, { value: 'published', label: '已发布' }]} /></Form.Item></Col>
          </Row>
          <Form.Item label="描述" name="description"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Modal>
      </Card>
    </div>
  )
}
