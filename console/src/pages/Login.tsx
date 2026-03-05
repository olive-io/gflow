import { useState } from 'react';
import { Form, Input, Button, Typography, App, Checkbox, Avatar } from 'antd';
import { 
  UserOutlined, 
  LockOutlined, 
  AppstoreOutlined,
  ArrowRightOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api';
import { useUserStore } from '../stores';

const { Title, Text, Paragraph } = Typography;

interface LoginForm {
  username: string;
  password: string;
}

const reviewers = [
  'https://lh3.googleusercontent.com/aida-public/AB6AXuDlPkjHShPRH4myH8eBwm6qR7O1jqaD7llJz6HfZps6HQXqGpmmotmsCPvAXXAzI76vi4mOFDwarhWhkDI8tmexti3AbtQm7Qa0C-fkZFkQK4d8FqFEqdRelwu5FSV9UQLO0Z_hziIJaOI0u27wui44kH_x-_sc8MDdrdG94nghQwT5VFw0lJUkGpaFm0UFZqxd2NCVbDCDZpiFpO1IGE0OLB5dqUfZHjoAWJU81PQXHHlLzdg74Ot1Muauqn2OncqOLfKIqqhnIJw',
  'https://lh3.googleusercontent.com/aida-public/AB6AXuAKqtOGm_mAyHNSs3xE7HUNoJBaEpaOAIgNKCty250DqGLS3ezVER_LTObGmeXdZaQwuebxDURD_kri9cjq1XGDid_WKodE95uKZuPGO4XCvxZTk5i88xu3-B7BVjK5oFdEH6HPuOIo0ze0RnXsRZO0TNJQpH_yI009sv904vEyd3f7h0S_qwfxzpaq0kX3mVMJFnE1PncFu4Q4Zhp052C42vi7vI38ai5KNCIrpW7EslWoxYFfnYJ1qH_JP2Ma9k3yTQbFyu6Er3w',
  'https://lh3.googleusercontent.com/aida-public/AB6AXuBPoaB9GkPtdrTmLv-OuYfYYeG4HiPq68ooUusrMeuE8GDggEiGc-REH7X-Js83cwstXPIkRowIxmxigtC5ASINsWqM7ehTJe2CRAGVLaOWUINihLpvhCeDr42HX_Pv3iCWuM2nzWERD1CL3E1lulxPQ7huURgz0BZxwnGsFaS3w5E_Wq4Q8V75IC4c6C5jUPt4JO2cEB7srkkx0QyO9VXB8FXFQP-gL8ThkLefZ04Bfiodpq3t0Lkr8JCaY-xtGf0bOfjNYAOS2r8',
];

const Login = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { setToken, setUser, setRole } = useUserStore();
  const { message } = App.useApp();

  const onFinish = async (values: LoginForm) => {
    setLoading(true);
    try {
      const response = await authApi.login(values.username, values.password);
      
      if (response.token?.text) {
        setToken(response.token);
        
        const selfResponse = await authApi.getSelf();
        if (selfResponse.user) {
          setUser(selfResponse.user);
          if (selfResponse.role) {
            setRole(selfResponse.role);
          }
        }
        
        navigate('/dashboard');
      } else {
        message.error('登录失败，未获取到令牌');
      }
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } }; message?: string };
      const errorMessage = err.response?.data?.message || err.message || '登录失败，请检查用户名和密码';
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="gflow-login-page">
      <div className="gflow-login-shell">
        <div className="gflow-login-hero">
          <div className="gflow-login-hero-pattern" />
          <div className="gflow-login-hero-glow one" />
          <div className="gflow-login-hero-glow two" />

          <div className="gflow-login-hero-content">
            <div className="gflow-login-hero-logo">
              <div className="gflow-login-hero-logo-mark">
                <AppstoreOutlined />
              </div>
              <Title level={2} className="gflow-login-hero-logo-text">GFlow</Title>
            </div>

            <Title level={1} className="gflow-login-hero-title">
              工作流引擎
              <br />
              <span>管理平台</span>
            </Title>

            <div className="gflow-login-hero-divider" />

            <Paragraph className="gflow-login-hero-desc">
              Go 语言构建的高性能 BPMN 解决方案，助力企业实现业务流程自动化与智能化。
            </Paragraph>
          </div>

          <div className="gflow-login-hero-proof">
            <Avatar.Group max={{ count: 3 }}>
              {reviewers.map((item) => (
                <Avatar key={item} src={item} size={36} />
              ))}
            </Avatar.Group>
            <Text className="gflow-login-hero-proof-text">
              <span>企业级可靠性</span>
              已有 500+ 企业接入生产环境
            </Text>
          </div>
        </div>

        <div className="gflow-login-panel">
          <div className="gflow-login-form-wrap">
            <div className="gflow-login-form-head">
              <Title level={4} className="gflow-login-form-title">欢迎登录</Title>
              <Text type="secondary" className="gflow-login-form-hint">
                请输入您的管理员账号访问系统控制台
              </Text>
            </div>

            <Form
              name="login"
              onFinish={onFinish}
              layout="vertical"
              size="large"
              requiredMark={false}
            >
              <Form.Item
                label="用户名"
                name="username"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input 
                  prefix={<UserOutlined />} 
                  placeholder="请输入用户名"
                />
              </Form.Item>

              <Form.Item
                label="密码"
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
              >
                <Input.Password 
                  prefix={<LockOutlined />} 
                  placeholder="请输入密码"
                  iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                />
              </Form.Item>

              <div className="gflow-login-actions-line">
                <Checkbox>记住我</Checkbox>
                <a href="#">忘记密码？</a>
              </div>

              <Form.Item className="gflow-login-form-submit-wrap">
                <Button 
                  type="primary" 
                  htmlType="submit" 
                  loading={loading} 
                  block
                  className="gflow-login-form-submit"
                >
                  <span>登录系统</span>
                  <ArrowRightOutlined />
                </Button>
              </Form.Item>
            </Form>

            <div className="gflow-login-tip">
              <div className="gflow-login-links">
                <a href="#">帮助中心</a>
                <a href="#">安全策略</a>
                <a href="#">服务协议</a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
