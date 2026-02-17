import { Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { AppShell, Group, Anchor, Container, Button, Text } from '@mantine/core'
import { Link } from 'react-router-dom'
import { useAuth } from './auth'
import RepoList from './pages/RepoList'
import RepoForm from './pages/RepoForm'
import RepoLogs from './pages/RepoLogs'
import Settings from './pages/Settings'
import Login from './pages/Login'
import PublicRepoLogs from './pages/PublicRepoLogs'

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { token } = useAuth()
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

function App() {
  const { token, user, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <AppShell header={{ height: 50 }} padding="md">
      <AppShell.Header>
        <Container size="md" h="100%">
          <Group h="100%" gap="lg" justify="space-between">
            <Group gap="lg">
              {token && <Anchor component={Link} to="/repos" fw={600} size="sm">Repositories</Anchor>}
              {token && isAdmin && <Anchor component={Link} to="/settings" fw={600} size="sm">Settings</Anchor>}
            </Group>
            {token && user && (
              <Group gap="sm">
                <Text size="sm" c="dimmed">{user.username} ({user.role})</Text>
                <Button size="compact-xs" variant="subtle" onClick={handleLogout}>Logout</Button>
              </Group>
            )}
          </Group>
        </Container>
      </AppShell.Header>
      <AppShell.Main>
        <Container size="md">
          <Routes>
            <Route path="/public/:repoName/:branch" element={<PublicRepoLogs />} />
            <Route path="/login" element={token ? <Navigate to="/repos" replace /> : <Login />} />
            <Route path="/" element={<Navigate to="/repos" replace />} />
            <Route path="/repos" element={<AuthGuard><RepoList /></AuthGuard>} />
            <Route path="/repos/new" element={<AuthGuard><RepoForm /></AuthGuard>} />
            <Route path="/repos/:id/edit" element={<AuthGuard><RepoForm /></AuthGuard>} />
            <Route path="/repos/:repoName/logs/:branch" element={<AuthGuard><RepoLogs /></AuthGuard>} />
            <Route path="/settings" element={<AuthGuard><Settings /></AuthGuard>} />
          </Routes>
        </Container>
      </AppShell.Main>
    </AppShell>
  )
}

export default App
