import { LockOutlined, UserOutlined } from '@ant-design/icons'
import { Button, Card, Form, Input, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import { setToken } from '../utils/auth'

type LoginForm = {
  username: string
  password: string
}

export default function LoginPage() {
  const navigate = useNavigate()

  const onFinish = async (values: LoginForm) => {
    const data = await api.login(values)
    setToken(data.token)
    navigate('/dashboard', { replace: true })
  }

  return (
    <div className="login-wrap">
      <Card className="login-card">
        <div className="login-avatar">👤</div>
        <Typography.Title level={2} className="login-title">在线学习管理平台</Typography.Title>
        <Form<LoginForm> className="login-form" onFinish={onFinish} initialValues={{ username: 'admin', password: 'admin123' }}>
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block className="login-btn">
            登录
          </Button>
        </Form>
        <Typography.Paragraph type="secondary" className="login-tip">测试账号：admin / admin123</Typography.Paragraph>
      </Card>
    </div>
  )
}
