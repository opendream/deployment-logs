import { useEffect, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { Title, Group, Stack, Text, Badge, Card, Loader, Center, Divider, Pagination, Container, AppShell } from '@mantine/core'
import { apiUrl } from '../auth'

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

export default function PublicRepoLogs() {
  const { repoName, branch } = useParams<{ repoName: string; branch: string }>()
  const [page, setPage] = useState(1)
  const [response, setResponse] = useState<PaginatedResponse | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchLogs = useCallback(() => {
    if (!repoName || !branch) return
    setLoading(true)
    const params = new URLSearchParams({ branch, page: String(page) })
    fetch(apiUrl(`/api/public/logs/repo/${encodeURIComponent(repoName)}?${params}`))
      .then(r => r.json())
      .then((data: PaginatedResponse) => setResponse(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [repoName, branch, page])

  useEffect(() => { fetchLogs() }, [fetchLogs])

  return (
    <AppShell header={{ height: 50 }} padding="md">
      <AppShell.Header>
        <Container size="md" h="100%">
          <Group h="100%" gap="lg">
            <Title order={4}>{repoName}</Title>
            <Badge size="lg" variant="filled">{branch}</Badge>
          </Group>
        </Container>
      </AppShell.Header>
      <AppShell.Main>
        <Container size="md">
          {loading && (
            <Center py="xl"><Loader size="sm" /></Center>
          )}

          {!loading && response && response.data.length === 0 && (
            <Card withBorder padding="lg">
              <Text c="dimmed" ta="center">No deployment logs yet for this branch.</Text>
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
        </Container>
      </AppShell.Main>
    </AppShell>
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
