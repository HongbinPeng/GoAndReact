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
      <Card title="系统登录" style={{ width: 380 }}>
        <Typography.Paragraph type="secondary">测试账号：admin / admin123</Typography.Paragraph>
        <Form<LoginForm> layout="vertical" onFinish={onFinish} initialValues={{ username: 'admin', password: 'admin123' }}>
          <Form.Item label="用户名" name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password />
          </Form.Item>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
        </Form>
      </Card>
    </div>
  )
}
