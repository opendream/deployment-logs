import { useEffect, useState } from 'react'
import { Title, Text, Stack, Card, TextInput, Button, Group, Code } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useAuth, apiHeaders, apiUrl, apiFetch } from '../auth'

export default function Settings() {
  const { token, user } = useAuth()
  const [apiKey, setApiKey] = useState('')
  const [maskedKey, setMaskedKey] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    apiFetch(apiUrl('/api/settings'), { headers: apiHeaders(token) })
      .then(r => r.json())
      .then(data => {
        setMaskedKey(data.app_key || '')
      })
      .catch(console.error)
  }, [token])

  const handleSave = async () => {
    setSaving(true)
    try {
      const resp = await apiFetch(apiUrl('/api/settings'), {
        method: 'PUT',
        headers: apiHeaders(token),
        body: JSON.stringify({ app_key: apiKey }),
      })
      if (!resp.ok) {
        const err = await resp.json()
        throw new Error(err.error || 'Failed to save')
      }
      notifications.show({ title: 'Saved', message: 'API key updated', color: 'green' })
      setApiKey('')
      // Refresh masked key
      const r = await apiFetch(apiUrl('/api/settings'), { headers: apiHeaders(token) })
      const data = await r.json()
      setMaskedKey(data.app_key || '')
    } catch (err: any) {
      notifications.show({ title: 'Error', message: err.message, color: 'red' })
    } finally {
      setSaving(false)
    }
  }

  const handleClear = async () => {
    setSaving(true)
    try {
      const resp = await apiFetch(apiUrl('/api/settings'), {
        method: 'PUT',
        headers: apiHeaders(token),
        body: JSON.stringify({ app_key: '' }),
      })
      if (!resp.ok) {
        const err = await resp.json()
        throw new Error(err.error || 'Failed to save')
      }
      notifications.show({ title: 'Saved', message: 'API key removed — trigger endpoint is now open', color: 'yellow' })
      setApiKey('')
      setMaskedKey('')
    } catch (err: any) {
      notifications.show({ title: 'Error', message: err.message, color: 'red' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <Title order={2} mb="md">Settings</Title>
      <Stack maw={500} gap="md">
        <Card withBorder>
          <Stack gap="xs">
            <Text size="sm"><Text span fw={600}>Username:</Text> {user?.username}</Text>
            <Text size="sm"><Text span fw={600}>Role:</Text> {user?.role}</Text>
          </Stack>
        </Card>

        <Card withBorder>
          <Stack gap="sm">
            <Text fw={600} size="sm">Trigger API Key</Text>
            <Text size="xs" c="dimmed">
              External systems can trigger log generation via <Code>POST /api/public/trigger</Code>.
              When an API key is set, requests must include <Code>X-API-Key</Code> header. If empty, the endpoint is open.
            </Text>
            {maskedKey && (
              <Text size="sm">Current key: <Code>{maskedKey}</Code></Text>
            )}
            {!maskedKey && (
              <Text size="sm" c="dimmed">No API key set — trigger endpoint is open.</Text>
            )}
            <TextInput
              placeholder="Enter new API key"
              value={apiKey}
              onChange={e => setApiKey(e.target.value)}
              styles={{ input: { fontFamily: 'monospace' } }}
            />
            <Group>
              <Button size="xs" onClick={handleSave} loading={saving} disabled={!apiKey.trim()}>
                Save Key
              </Button>
              {maskedKey && (
                <Button size="xs" variant="light" color="red" onClick={handleClear} loading={saving}>
                  Remove Key
                </Button>
              )}
            </Group>
          </Stack>
        </Card>
      </Stack>
    </>
  )
}
