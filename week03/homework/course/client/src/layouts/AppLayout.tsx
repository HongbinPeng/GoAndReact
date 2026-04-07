import { BookOutlined, DashboardOutlined, DownOutlined, FileTextOutlined, LogoutOutlined, MenuOutlined, RobotOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons'
import { Dropdown, Layout, Menu, type MenuProps } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { clearToken } from '../utils/auth'

const { Header, Sider, Content } = Layout

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()

  const selected = location.pathname.split('/')[1] || 'dashboard'
  const userMenu: MenuProps['items'] = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        clearToken()
        navigate('/login', { replace: true })
      },
    },
  ]

  return (
    <Layout className="app-shell">
      <Sider theme="light" width={168} className="app-sider">
        <div className="logo">🎓 学习管理平台</div>
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[selected]}
          className="side-menu"
          items={[
            { key: 'dashboard', icon: <DashboardOutlined />, label: '工作台' },
            { key: 'courses', icon: <BookOutlined />, label: '课程管理' },
            { key: 'students', icon: <TeamOutlined />, label: '学生管理' },
            { key: 'summary', icon: <FileTextOutlined />, label: '学习总结' },
            { key: 'ai-chat', icon: <RobotOutlined />, label: 'AI 对话' },
          ]}
          onClick={({ key }) => navigate(`/${key}`)}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <div className="app-title"><MenuOutlined /></div>
          <Dropdown menu={{ items: userMenu }} trigger={['click']}>
            <div className="admin-box">
              <UserOutlined />
              <span>管理员</span>
              <DownOutlined style={{ fontSize: 10 }} />
            </div>
          </Dropdown>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
