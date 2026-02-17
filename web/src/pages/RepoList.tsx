import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Title, Group, Button, Table, ActionIcon, Text, Badge, Anchor } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useAuth, apiHeaders, apiUrl, apiFetch } from '../auth'

interface Config {
  id: number
  name: string
  provider: string
  default_branch: string
  branches: string
}

interface BranchConfigEntry {
  pattern: string
  environment: string
  date_strategy: string
}

function parseEnvironments(config: Config): string[] {
  if (config.branches) {
    try {
      const parsed = JSON.parse(config.branches)
      if (Array.isArray(parsed) && parsed.length > 0) {
        // New format: array of objects
        if (typeof parsed[0] === 'object' && parsed[0].pattern) {
          const envs = (parsed as BranchConfigEntry[]).map(b => b.environment)
          return [...new Set(envs)]
        }
        // Legacy format: array of strings
        return parsed
      }
    } catch { /* fall through */ }
  }
  return [config.default_branch]
}

export default function RepoList() {
  const { token, isAdmin } = useAuth()
  const [configs, setConfigs] = useState<Config[]>([])

  const fetchConfigs = () => {
    apiFetch(apiUrl('/api/repos'), { headers: apiHeaders(token) }, { noRedirect: true })
      .then(r => r.json())
      .then(setConfigs)
      .catch(console.error)
  }

  useEffect(() => { fetchConfigs() }, [])

  const handleDelete = (id: number) => {
    if (!confirm('Delete this repository?')) return
    apiFetch(apiUrl(`/api/repos/${id}`), { method: 'DELETE', headers: apiHeaders(token) })
      .then(() => {
        notifications.show({ title: 'Deleted', message: 'Repository removed', color: 'red' })
        fetchConfigs()
      })
      .catch(console.error)
  }

  return (
    <>
      <Group justify="space-between" mb="md">
        <Title order={2}>Repositories</Title>
        {isAdmin && <Button component={Link} to="/repos/new" size="sm">Add Repository</Button>}
      </Group>

      {configs.length === 0 ? (
        <Text c="dimmed">No repositories yet.</Text>
      ) : (
        <Table striped highlightOnHover withTableBorder>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Provider</Table.Th>
              <Table.Th>Environments</Table.Th>
              {isAdmin && <Table.Th w={100}>Actions</Table.Th>}
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {configs.map(c => {
              const environments = parseEnvironments(c)
              return (
                <Table.Tr key={c.id}>
                  <Table.Td>
                    <Text size="sm" fw={500} style={{ fontFamily: 'monospace' }}>{c.name}</Text>
                  </Table.Td>
                  <Table.Td>
                    <Badge size="sm" variant="light">{c.provider}</Badge>
                  </Table.Td>
                  <Table.Td>
                    <Group gap={6}>
                      {environments.map(env => (
                        <Anchor
                          key={env}
                          component={Link}
                          to={`/repos/${c.name}/logs/${encodeURIComponent(env)}`}
                          size="sm"
                          underline="never"
                        >
                          <Badge variant="outline" size="sm" style={{ cursor: 'pointer' }}>
                            {env}
                          </Badge>
                        </Anchor>
                      ))}
                    </Group>
                  </Table.Td>
                  {isAdmin && (
                    <Table.Td>
                      <Group gap="xs">
                        <Button component={Link} to={`/repos/${c.id}/edit`} size="compact-xs" variant="light">Edit</Button>
                        <ActionIcon color="red" variant="light" size="sm" onClick={() => handleDelete(c.id)}>
                          <span style={{ fontSize: 14 }}>x</span>
                        </ActionIcon>
                      </Group>
                    </Table.Td>
                  )}
                </Table.Tr>
              )
            })}
          </Table.Tbody>
        </Table>
      )}
    </>
  )
}
