import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { Title, Group, Stack, Text, Badge, Card, Loader, Center, Divider, Pagination, Button, Breadcrumbs, Anchor, Modal, TextInput, Menu, ActionIcon } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import { IconSettings } from '@tabler/icons-react'
import { useAuth, apiHeaders, apiUrl, apiFetch } from '../auth'

interface LogItem {
  id: number
  title: string
  description: string
  pr_number?: number
  category: string
  commit_sha: string
  merged_at: string
}

interface DateGroup {
  date: string
  items: LogItem[]
}

interface PaginatedResponse {
  page: number
  per_page: number
  total_pages: number
  total_items: number
  data: DateGroup[]
}

interface BranchConfigEntry {
  pattern: string
  environment: string
  date_strategy: string
}

const categoryColors: Record<string, string> = {
  feat: 'blue',
  fix: 'red',
  hotfix: 'orange',
  chore: 'gray',
  refactor: 'violet',
}

function formatDate(dateStr: string): string {
  const [y, m, d] = dateStr.split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

function hasGlobPattern(pattern: string): boolean {
  return pattern.includes('*') || pattern.includes('?')
}

export default function RepoLogs() {
  const { token, isAdmin } = useAuth()
  const { repoName, branch } = useParams<{ repoName: string; branch: string }>()
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [response, setResponse] = useState<PaginatedResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [generating, setGenerating] = useState(false)
  const [environments, setEnvironments] = useState<string[]>([])
  const [branchConfigs, setBranchConfigs] = useState<BranchConfigEntry[]>([])
  const [clearModalOpen, setClearModalOpen] = useState(false)
  const [clearConfirmText, setClearConfirmText] = useState('')
  const [clearing, setClearing] = useState(false)
  const [generateModalOpen, setGenerateModalOpen] = useState(false)
  const [generateBranch, setGenerateBranch] = useState('')

  const fetchLogs = useCallback(() => {
    if (!repoName || !branch) return
    setLoading(true)
    const params = new URLSearchParams({ branch, page: String(page) })
    apiFetch(apiUrl(`/api/logs/repo/${encodeURIComponent(repoName)}?${params}`), { headers: apiHeaders(token) }, { noRedirect: true })
      .then(r => r.json())
      .then((data: PaginatedResponse) => setResponse(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [repoName, branch, page, token])

  useEffect(() => { fetchLogs() }, [fetchLogs])

  useEffect(() => {
    if (!repoName) return
    apiFetch(apiUrl('/api/repos'), { headers: apiHeaders(token) }, { noRedirect: true })
      .then(r => r.json())
      .then((repos: any[]) => {
        const repo = repos.find((r: any) => r.name === repoName)
        if (!repo) return
        if (repo.branches) {
          try {
            const parsed = JSON.parse(repo.branches)
            if (Array.isArray(parsed) && parsed.length > 0) {
              if (typeof parsed[0] === 'object' && parsed[0].pattern) {
                const configs = parsed as BranchConfigEntry[]
                setBranchConfigs(configs)
                const envs = configs.map(b => b.environment)
                setEnvironments([...new Set(envs)])
                return
              }
              // Legacy format
              setEnvironments(parsed)
              setBranchConfigs([])
              return
            }
          } catch { /* ignore */ }
        }
        setEnvironments([repo.default_branch])
        setBranchConfigs([])
      })
      .catch(console.error)
  }, [repoName, token])

  // Find if current environment uses a glob pattern
  const currentConfig = branchConfigs.find(c => c.environment === branch)
  const isPatternEnv = currentConfig ? hasGlobPattern(currentConfig.pattern) : false

  const triggerGenerate = async (targetBranch: string) => {
    if (!repoName) return
    setGenerating(true)
    try {
      const resp = await apiFetch(apiUrl('/api/trigger'), {
        method: 'POST',
        headers: apiHeaders(token),
        body: JSON.stringify({ repo_name: repoName, branch: targetBranch }),
      })
      if (!resp.ok) {
        const err = await resp.json()
        throw new Error(err.error || 'Generation failed')
      }
      const log = await resp.json()
      const itemCount = log.LogItems?.length ?? log.log_items?.length ?? 0
      notifications.show({
        title: 'Generated',
        message: `${itemCount} item${itemCount !== 1 ? 's' : ''} found`,
        color: 'green',
      })
      setGenerateModalOpen(false)
      setGenerateBranch('')
      fetchLogs()
    } catch (err: any) {
      notifications.show({ title: 'Error', message: err.message, color: 'red' })
    } finally {
      setGenerating(false)
    }
  }

  const handleGenerate = () => {
    if (isPatternEnv && currentConfig) {
      // Open modal for pattern-based environments
      const prefix = currentConfig.pattern.replace(/\*.*$/, '')
      setGenerateBranch(prefix)
      setGenerateModalOpen(true)
    } else {
      // Direct trigger for non-pattern environments
      triggerGenerate(branch!)
    }
  }

  const confirmPhrase = `${repoName}/${branch}`

  const handleClear = async () => {
    if (!repoName || !branch) return
    setClearing(true)
    try {
      const params = new URLSearchParams({ branch })
      const resp = await apiFetch(apiUrl(`/api/logs/repo/${encodeURIComponent(repoName)}?${params}`), {
        method: 'DELETE',
        headers: apiHeaders(token),
      })
      if (!resp.ok) {
        const err = await resp.json()
        throw new Error(err.error || 'Clear failed')
      }
      const result = await resp.json()
      notifications.show({
        title: 'Cleared',
        message: result.message,
        color: 'green',
      })
      setClearModalOpen(false)
      setClearConfirmText('')
      fetchLogs()
    } catch (err: any) {
      notifications.show({ title: 'Error', message: err.message, color: 'red' })
    } finally {
      setClearing(false)
    }
  }

  return (
    <>
      {/* Generate modal for pattern environments */}
      <Modal
        opened={generateModalOpen}
        onClose={() => { setGenerateModalOpen(false); setGenerateBranch('') }}
        title="Generate Logs"
        centered
      >
        <Stack gap="sm">
          <Text size="sm">
            Enter the exact branch name to generate logs for the <Text span fw={700}>{branch}</Text> environment.
          </Text>
          <TextInput
            label="Branch name"
            placeholder={currentConfig?.pattern || ''}
            value={generateBranch}
            onChange={e => setGenerateBranch(e.target.value)}
            styles={{ input: { fontFamily: 'monospace' } }}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={() => { setGenerateModalOpen(false); setGenerateBranch('') }}>
              Cancel
            </Button>
            <Button
              disabled={!generateBranch.trim()}
              loading={generating}
              onClick={() => triggerGenerate(generateBranch.trim())}
            >
              Generate
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Clear modal */}
      <Modal
        opened={clearModalOpen}
        onClose={() => { setClearModalOpen(false); setClearConfirmText('') }}
        title="Clear All Logs"
        centered
      >
        <Stack gap="sm">
          <Text size="sm">
            This will permanently delete all deployment logs for <Text span fw={700}>{repoName}</Text> on environment <Text span fw={700}>{branch}</Text>.
          </Text>
          <Text size="sm" c="dimmed">
            Type <Text span ff="monospace" fw={700} c="red">{confirmPhrase}</Text> to confirm.
          </Text>
          <TextInput
            placeholder={confirmPhrase}
            value={clearConfirmText}
            onChange={e => setClearConfirmText(e.target.value)}
            styles={{ input: { fontFamily: 'monospace' } }}
          />
          <Group justify="flex-end">
            <Button variant="default" onClick={() => { setClearModalOpen(false); setClearConfirmText('') }}>
              Cancel
            </Button>
            <Button
              color="red"
              disabled={clearConfirmText !== confirmPhrase}
              loading={clearing}
              onClick={handleClear}
            >
              Clear Logs
            </Button>
          </Group>
        </Stack>
      </Modal>

      <Breadcrumbs mb="sm">
        <Anchor component={Link} to="/repos" size="sm">Repositories</Anchor>
        <Text size="sm" style={{ fontFamily: 'monospace' }}>{repoName}</Text>
        <Text size="sm">{branch}</Text>
      </Breadcrumbs>

      <Group justify="space-between" mb="md" align="center">
        <Group gap="sm">
          <Title order={2}>{repoName}</Title>
          {environments.map(env => (
            <Badge
              key={env}
              size="lg"
              variant={env === branch ? 'filled' : 'light'}
              style={{ cursor: env === branch ? 'default' : 'pointer' }}
              onClick={() => {
                if (env !== branch) {
                  setPage(1)
                  navigate(`/repos/${repoName}/logs/${encodeURIComponent(env)}`)
                }
              }}
            >
              {env}
            </Badge>
          ))}
          {branch && !environments.includes(branch) && environments.length > 0 && (
            <Badge size="lg" variant="filled">{branch}</Badge>
          )}
        </Group>
        {isAdmin && (
          <Menu position="bottom-end">
            <Menu.Target>
              <ActionIcon variant="subtle" size="lg">
                <IconSettings size={18} />
              </ActionIcon>
            </Menu.Target>
            <Menu.Dropdown>
              <Menu.Item onClick={handleGenerate}>Generate</Menu.Item>
              {response && response.total_items > 0 && (
                <Menu.Item color="red" onClick={() => setClearModalOpen(true)}>Clear Logs</Menu.Item>
              )}
            </Menu.Dropdown>
          </Menu>
        )}
      </Group>

      {loading && (
        <Center py="xl"><Loader size="sm" /></Center>
      )}

      {!loading && response && response.data.length === 0 && (
        <Card withBorder padding="lg">
          <Stack align="center" gap="sm">
            <Text c="dimmed">No deployment logs yet for this environment.</Text>
            {isAdmin && (
              <Button onClick={handleGenerate} loading={generating} variant="light">
                Generate Now
              </Button>
            )}
          </Stack>
        </Card>
      )}

      {!loading && response && response.data.length > 0 && (
        <Stack gap="md">
          <Text size="sm" c="dimmed">{response.total_items} total items across {response.total_pages} page{response.total_pages !== 1 ? 's' : ''}</Text>

          {response.data.map(group => (
            <div key={group.date}>
              <Divider
                label={
                  <Group gap="xs">
                    <Text fw={600} size="sm">{formatDate(group.date)}</Text>
                    <Badge size="xs" variant="light" color="gray">{group.items.length}</Badge>
                  </Group>
                }
                labelPosition="left"
                mb="sm"
              />
              <Stack gap="xs">
                {group.items.map(item => (
                  <ItemCard key={item.id} item={item} />
                ))}
              </Stack>
            </div>
          ))}

          {response.total_pages > 1 && (
            <Center mt="md">
              <Pagination
                total={response.total_pages}
                value={page}
                onChange={setPage}
              />
            </Center>
          )}
        </Stack>
      )}
    </>
  )
}

function ItemCard({ item }: { item: LogItem }) {
  const [opened, setOpened] = useState(false)
  const hasDetail = Boolean(item.description)

  return (
    <Card
      withBorder
      padding="xs"
      style={hasDetail ? { cursor: 'pointer' } : undefined}
      onClick={hasDetail ? () => setOpened(o => !o) : undefined}
    >
      <Group justify="space-between" wrap="nowrap">
        <Group gap="xs" wrap="nowrap" style={{ minWidth: 0 }}>
          <Badge
            color={categoryColors[item.category] || 'gray'}
            size="xs"
            variant="light"
            style={{ flexShrink: 0 }}
          >
            {item.category}
          </Badge>
          {item.pr_number && (
            <Badge size="xs" variant="outline" style={{ flexShrink: 0 }}>#{item.pr_number}</Badge>
          )}
          <Text size="sm" truncate>{item.title}</Text>
        </Group>
        <Text size="xs" c="dimmed" style={{ fontFamily: 'monospace', flexShrink: 0 }}>
          {item.commit_sha.slice(0, 8)}
        </Text>
      </Group>
      {opened && item.description && (
        <Text size="xs" c="dimmed" style={{ whiteSpace: 'pre-wrap' }} mt="xs">
          {item.description}
        </Text>
      )}
    </Card>
  )
}
