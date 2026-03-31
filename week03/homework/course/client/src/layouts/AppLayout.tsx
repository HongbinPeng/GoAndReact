import { BookOutlined, DashboardOutlined, FileTextOutlined, TeamOutlined, LogoutOutlined } from '@ant-design/icons'
import { Button, Layout, Menu, Typography } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { clearToken } from '../utils/auth'

const { Header, Sider, Content } = Layout

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()

  const selected = location.pathname.split('/')[1] || 'dashboard'

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="light">
        <div className="logo">课程管理系统</div>
        <Menu
          mode="inline"
          selectedKeys={[selected]}
          items={[
            { key: 'dashboard', icon: <DashboardOutlined />, label: '工作台' },
            { key: 'courses', icon: <BookOutlined />, label: '课程管理' },
            { key: 'students', icon: <TeamOutlined />, label: '学生管理' },
            { key: 'summary', icon: <FileTextOutlined />, label: '学习总结' },
          ]}
          onClick={({ key }) => navigate(`/${key}`)}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <Typography.Text style={{ color: '#fff' }}>欢迎使用课程管理系统</Typography.Text>
          <Button
            icon={<LogoutOutlined />}
            onClick={() => {
              clearToken()
              navigate('/login', { replace: true })
            }}
          >
            退出登录
          </Button>
        </Header>
        <Content style={{ padding: 16 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
