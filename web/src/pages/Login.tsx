import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Title, TextInput, PasswordInput, Button, Stack, Card, Center, Text } from '@mantine/core'
import { useAuth, apiUrl } from '../auth'

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const resp = await fetch(apiUrl('/api/auth/login'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      const data = await resp.json()
      if (!resp.ok) {
        setError(data.error || 'Login failed')
        return
      }
      login(data.token, data.user)
      navigate('/repos', { replace: true })
    } catch {
      setError('Network error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Center mih="60vh">
      <Card withBorder shadow="sm" p="xl" maw={400} w="100%">
        <Title order={3} mb="md" ta="center">Login</Title>
        <form onSubmit={handleSubmit}>
          <Stack gap="sm">
            <TextInput label="Username" value={username} onChange={e => setUsername(e.target.value)} required />
            <PasswordInput label="Password" value={password} onChange={e => setPassword(e.target.value)} required />
            {error && <Text c="red" size="sm">{error}</Text>}
            <Button type="submit" loading={loading} fullWidth>Login</Button>
          </Stack>
        </form>
      </Card>
    </Center>
  )
}
