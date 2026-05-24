<template>
  <div class="flex flex-col h-full bg-background">
    <header class="flex items-center gap-3 px-6 py-4 border-b">
      <Button
        variant="ghost"
        size="sm"
        @click="router.push({ name: 'team-workspace', params: { teamId } })"
      >
        <ArrowLeft class="size-4" />
      </Button>
      <div class="flex-1">
        <div class="text-xs text-muted-foreground">
          {{ team?.name ?? '' }} #{{ issue?.number ?? '' }}
        </div>
        <div class="text-lg font-semibold">
          {{ issue?.title ?? t('teams.loading') }}
        </div>
      </div>
      <Badge
        v-if="issue"
        :variant="badgeVariantForStatus(issue.status)"
      >
        {{ issue.status }}
      </Badge>
    </header>

    <div class="flex-1 overflow-auto px-6 py-4 space-y-4">
      <Card v-if="issue?.description">
        <CardContent class="pt-6 whitespace-pre-wrap text-sm">
          {{ issue.description }}
        </CardContent>
      </Card>

      <Card v-if="botSessions.length > 0">
        <CardContent class="pt-6 space-y-2 text-sm">
          <div class="text-xs uppercase tracking-wide text-muted-foreground mb-1">
            {{ t('teams.botSessions') }}
          </div>
          <p class="text-xs text-muted-foreground mb-2">
            {{ t('teams.botSessionsHint') }}
          </p>
          <ul class="space-y-1.5">
            <li
              v-for="s in botSessions"
              :key="s.bot_id + ':' + s.session_id"
              class="flex items-center justify-between"
            >
              <span class="font-medium">{{ s.bot_name || s.bot_id }}</span>
              <Button
                variant="outline"
                size="sm"
                @click="openBotSession(s.bot_id, s.session_id)"
              >
                <MessageSquare class="size-3.5 mr-1" /> {{ t('teams.openSession') }}
              </Button>
            </li>
          </ul>
        </CardContent>
      </Card>

      <div class="space-y-3">
        <Card
          v-for="cmt in comments"
          :key="cmt.id"
        >
          <CardContent class="pt-6 space-y-2 text-sm">
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              <span class="font-medium">{{ cmt.author?.type }}</span>
              <span>·</span>
              <span>{{ formatDate(cmt.created_at) }}</span>
            </div>
            <div class="whitespace-pre-wrap">
              {{ cmt.content }}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

    <footer class="border-t px-6 py-4 space-y-2">
      <Textarea
        v-model="newComment"
        :placeholder="t('teams.commentPlaceholder')"
        rows="3"
      />
      <div class="flex items-center justify-end gap-2">
        <Button
          :disabled="!newComment.trim() || posting"
          @click="submitComment"
        >
          <Send class="size-4 mr-1" /> {{ t('teams.postComment') }}
        </Button>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import { ArrowLeft, MessageSquare, Send } from 'lucide-vue-next'
import {
  Badge,
  Button,
  Card,
  CardContent,
  Textarea,
} from '@memohai/ui'
import {
  getBots,
  getTeamsByTeamId,
  getTeamsByTeamIdIssuesByIssueId,
  getTeamsByTeamIdIssuesByIssueIdComments,
  getTeamsByTeamIdIssuesByIssueIdHandoffs,
  postTeamsByTeamIdIssuesByIssueIdComments,
} from '@memohai/sdk'
import { resolveApiErrorMessage } from '@/utils/api-error'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const queryCache = useQueryCache()

const teamId = computed(() => String(route.params.teamId ?? ''))
const issueId = computed(() => String(route.params.issueId ?? ''))

const { data: teamData } = useQuery({
  key: () => ['team', teamId.value],
  query: async () => {
    const { data, error } = await getTeamsByTeamId({ path: { team_id: teamId.value } })
    if (error) throw error
    return data
  },
  enabled: () => !!teamId.value,
})

const { data: issueData } = useQuery({
  key: () => ['team', teamId.value, 'issue', issueId.value],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdIssuesByIssueId({
      path: { team_id: teamId.value, issue_id: issueId.value },
    })
    if (error) throw error
    return data
  },
  enabled: () => !!teamId.value && !!issueId.value,
})

const { data: commentsData, refetch: refetchComments } = useQuery({
  key: () => ['team', teamId.value, 'issue', issueId.value, 'comments'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdIssuesByIssueIdComments({
      path: { team_id: teamId.value, issue_id: issueId.value },
    })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value && !!issueId.value,
})

const { data: handoffsData } = useQuery({
  key: () => ['team', teamId.value, 'issue', issueId.value, 'handoffs'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdIssuesByIssueIdHandoffs({
      path: { team_id: teamId.value, issue_id: issueId.value },
    })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value && !!issueId.value,
})

const { data: botListData } = useQuery({
  key: () => ['bots-for-issue', teamId.value, issueId.value],
  query: async () => {
    const { data, error } = await getBots()
    if (error) throw error
    return data?.items ?? []
  },
})

const team = computed(() => teamData.value)
const issue = computed(() => issueData.value)
const comments = computed(() => commentsData.value ?? [])

// Distinct (bot_id, session_id) pairs for handoffs that have a session
// attached. Each pair is one transcript a user can open in chat. Bots
// without a session (e.g. handoffs that errored before persistence)
// are filtered out — there is nothing to show.
const botSessions = computed(() => {
  const handoffs = handoffsData.value ?? []
  const bots = botListData.value ?? []
  const nameByID = new Map<string, string>()
  for (const b of bots) {
    if (b.id) nameByID.set(b.id, b.display_name ?? '')
  }
  const seen = new Set<string>()
  const out: Array<{ bot_id: string; bot_name: string; session_id: string }> = []
  for (const h of handoffs) {
    const botID = h.to_bot_id ?? ''
    const sid = h.target_session_id ?? ''
    if (!botID || !sid) continue
    const key = `${botID}:${sid}`
    if (seen.has(key)) continue
    seen.add(key)
    out.push({ bot_id: botID, bot_name: nameByID.get(botID) ?? '', session_id: sid })
  }
  return out
})

function openBotSession(botID: string, sessionID: string) {
  if (!botID || !sessionID) return
  router.push({ name: 'chat', params: { botId: botID }, query: { session: sessionID } })
}

const newComment = ref('')

const { mutate: doPost, status: postStatus } = useMutation({
  mutation: async () => {
    const { data, error } = await postTeamsByTeamIdIssuesByIssueIdComments({
      path: { team_id: teamId.value, issue_id: issueId.value },
      body: { content: newComment.value.trim() },
    })
    if (error) throw error
    return data
  },
  onSuccess: () => {
    newComment.value = ''
    void refetchComments()
    void queryCache.invalidateQueries({ key: ['team', teamId.value, 'issue', issueId.value] })
  },
  onError: (err) => {
    toast.error(resolveApiErrorMessage(err, t('teams.commentFailed')))
  },
})

const posting = computed(() => postStatus.value === 'loading')

function submitComment() {
  if (!newComment.value.trim()) return
  doPost()
}

function formatDate(value: string | undefined): string {
  if (!value) return ''
  try {
    return new Date(value).toLocaleString()
  }
  catch {
    return value
  }
}

function badgeVariantForStatus(status: string | undefined): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'in_progress':
    case 'review':
      return 'default'
    case 'blocked':
      return 'destructive'
    case 'done':
    case 'cancelled':
      return 'outline'
    default:
      return 'secondary'
  }
}
</script>
