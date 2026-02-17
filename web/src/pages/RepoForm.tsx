import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Title, TextInput, Select, Textarea, Button, Group, Stack, Table, ActionIcon, Text } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { useAuth, apiHeaders, apiUrl, apiFetch } from '../auth'

const DATE_STRATEGY_OPTIONS = [
  { value: 'merge_date', label: 'PR/Commit Date' },
  { value: 'deploy_date', label: 'Deploy Date' },
]

const DEFAULT_BRANCHES: BranchRow[] = [
  { pattern: 'develop', environment: 'develop', dateStrategy: 'merge_date' },
  { pattern: 'release/*', environment: 'uat', dateStrategy: 'deploy_date' },
  { pattern: 'main', environment: 'production', dateStrategy: 'deploy_date' },
]

interface BranchRow {
  pattern: string
  environment: string
  dateStrategy: string
}

export default function RepoForm() {
  const { token } = useAuth()
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = Boolean(id)

  const [form, setForm] = useState({
    name: '',
    repo_url: '',
    provider: 'github',
    ssh_key: '',
    default_branch: 'main',
  })
  const [branches, setBranches] = useState<BranchRow[]>(DEFAULT_BRANCHES)
  const [hasExistingKey, setHasExistingKey] = useState(false)

  useEffect(() => {
    if (!id) return
    apiFetch(apiUrl(`/api/repos/${id}`), { headers: apiHeaders(token) })
      .then(r => r.json())
      .then(data => {
        setForm({
          name: data.name || '',
          repo_url: data.repo_url || '',
          provider: data.provider || 'github',
          ssh_key: '',
          default_branch: data.default_branch || 'main',
        })
        setHasExistingKey(true)

        if (data.branches) {
          try {
            const parsed = JSON.parse(data.branches)
            if (Array.isArray(parsed) && parsed.length > 0) {
              // New format: array of objects with pattern/environment/date_strategy
              if (typeof parsed[0] === 'object' && parsed[0].pattern) {
                setBranches(parsed.map((b: any) => ({
                  pattern: b.pattern || '',
                  environment: b.environment || b.pattern || '',
                  dateStrategy: b.date_strategy || 'merge_date',
                })))
              } else {
                // Legacy format: array of strings
                let strategy: Record<string, string> = {}
                if (data.branch_date_strategy) {
                  try { strategy = JSON.parse(data.branch_date_strategy) } catch { /* ignore */ }
                }
                setBranches(parsed.map((name: string) => ({
                  pattern: name,
                  environment: name,
                  dateStrategy: strategy[name] || 'merge_date',
                })))
              }
            }
          } catch { /* ignore */ }
        }
      })
      .catch(console.error)
  }, [id])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const validBranches = branches.filter(b => b.pattern.trim())
    const branchConfigs = validBranches.map(b => ({
      pattern: b.pattern.trim(),
      environment: b.environment.trim() || b.pattern.trim(),
      date_strategy: b.dateStrategy,
    }))

    const body: Record<string, string> = {
      name: form.name,
      repo_url: form.repo_url,
      provider: form.provider,
      default_branch: form.default_branch,
      branches: branchConfigs.length > 0 ? JSON.stringify(branchConfigs) : '',
    }
    if (form.ssh_key) {
      body.ssh_key = form.ssh_key
    }

    const url = isEdit ? apiUrl(`/api/repos/${id}`) : apiUrl('/api/repos')
    const method = isEdit ? 'PUT' : 'POST'
    fetch(url, { method, headers: apiHeaders(token), body: JSON.stringify(body) })
      .then(r => {
        if (r.ok) {
          notifications.show({ title: 'Saved', message: `Repository ${isEdit ? 'updated' : 'created'}`, color: 'green' })
          navigate('/repos')
        } else {
          r.json().then(err => notifications.show({ title: 'Error', message: err.error, color: 'red' }))
        }
      })
      .catch(console.error)
  }

  const set = (field: string) => (val: string | null) =>
    setForm(f => ({ ...f, [field]: val ?? '' }))

  const setInput = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
    setForm(f => ({ ...f, [field]: e.target.value }))

  const setName = (e: React.ChangeEvent<HTMLInputElement>) => {
    const slug = e.target.value
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9\-_]/g, '')
    setForm(f => ({ ...f, name: slug }))
  }

  const nameError = form.name && !/^[a-z0-9][a-z0-9\-_]*$/.test(form.name)
    ? 'Lowercase letters, numbers, hyphens and underscores only'
    : undefined

  const addBranch = () => {
    setBranches(prev => [...prev, { pattern: '', environment: '', dateStrategy: 'merge_date' }])
  }

  const removeBranch = (index: number) => {
    setBranches(prev => prev.filter((_, i) => i !== index))
  }

  const updateBranch = (index: number, field: keyof BranchRow, value: string) => {
    setBranches(prev => prev.map((b, i) => i === index ? { ...b, [field]: value } : b))
  }

  return (
    <>
      <Title order={2} mb="md">{isEdit ? 'Edit Repository' : 'New Repository'}</Title>
      <form onSubmit={handleSubmit}>
        <Stack maw={700} gap="sm">
          <TextInput
            label="Name"
            value={form.name}
            onChange={setName}
            required
            error={nameError}
            placeholder="my-repo-name"
            description="Lowercase, hyphens and underscores only"
          />
          <TextInput
            label="Repo URL"
            value={form.repo_url}
            onChange={setInput('repo_url')}
            required
            placeholder="git@github.com:org/repo.git"
          />
          <Group grow>
            <Select
              label="Provider"
              data={[{ value: 'github', label: 'GitHub' }, { value: 'gitlab', label: 'GitLab' }]}
              value={form.provider}
              onChange={set('provider')}
            />
            <TextInput
              label="Default Branch"
              value={form.default_branch}
              onChange={setInput('default_branch')}
            />
          </Group>
          <Textarea
            label="SSH Key (private key)"
            value={form.ssh_key}
            onChange={setInput('ssh_key')}
            minRows={4}
            autosize
            placeholder={isEdit && hasExistingKey
              ? 'Leave empty to keep existing key, or paste new key to replace'
              : '-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----'}
            styles={{ input: { fontFamily: 'monospace', fontSize: '0.8rem' } }}
            description={isEdit && hasExistingKey
              ? 'SSH key is stored securely. Leave empty to keep current key, or paste a new key to replace it.'
              : 'Paste your SSH private key for git authentication'}
          />

          <div>
            <Group justify="space-between" mb="xs">
              <Text fw={500} size="sm">Branches</Text>
              <Button size="compact-xs" variant="light" onClick={addBranch}>+ Add Branch</Button>
            </Group>
            <Table withTableBorder withColumnBorders>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Pattern</Table.Th>
                  <Table.Th w={140}>Environment</Table.Th>
                  <Table.Th w={180}>Date Strategy</Table.Th>
                  <Table.Th w={50} />
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {branches.length === 0 && (
                  <Table.Tr>
                    <Table.Td colSpan={4}>
                      <Text size="sm" c="dimmed" ta="center" py="xs">
                        No branches configured. Click "+ Add Branch" to add one.
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                )}
                {branches.map((branch, i) => (
                  <Table.Tr key={i}>
                    <Table.Td>
                      <TextInput
                        size="xs"
                        variant="unstyled"
                        value={branch.pattern}
                        onChange={e => updateBranch(i, 'pattern', e.target.value)}
                        placeholder="release/*"
                        styles={{ input: { fontFamily: 'monospace' } }}
                      />
                    </Table.Td>
                    <Table.Td>
                      <TextInput
                        size="xs"
                        variant="unstyled"
                        value={branch.environment}
                        onChange={e => updateBranch(i, 'environment', e.target.value)}
                        placeholder="uat"
                        styles={{ input: { fontFamily: 'monospace' } }}
                      />
                    </Table.Td>
                    <Table.Td>
                      <Select
                        size="xs"
                        variant="unstyled"
                        data={DATE_STRATEGY_OPTIONS}
                        value={branch.dateStrategy}
                        onChange={val => updateBranch(i, 'dateStrategy', val || 'merge_date')}
                      />
                    </Table.Td>
                    <Table.Td>
                      <ActionIcon
                        color="red"
                        variant="subtle"
                        size="sm"
                        onClick={() => removeBranch(i)}
                      >
                        <span style={{ fontSize: 14 }}>x</span>
                      </ActionIcon>
                    </Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </div>

          <Group mt="sm">
            <Button type="submit">Save</Button>
            <Button variant="default" onClick={() => navigate('/repos')}>Cancel</Button>
          </Group>
        </Stack>
      </form>
    </>
  )
}
