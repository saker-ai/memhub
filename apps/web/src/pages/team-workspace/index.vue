<template>
  <div class="flex flex-col h-full bg-background">
    <header class="flex items-center justify-between px-6 py-4 border-b">
      <div>
        <div class="text-xs text-muted-foreground">
          {{ t('teams.title') }}
        </div>
        <div class="text-lg font-semibold">
          {{ team?.name ?? t('teams.loading') }}
        </div>
      </div>
      <Button @click="showCreate = true">
        <Plus class="size-4 mr-1" /> {{ t('teams.newIssue') }}
      </Button>
    </header>

    <div class="flex-1 overflow-auto px-6 py-4">
      <div
        v-if="issues.length === 0"
        class="text-sm text-muted-foreground"
      >
        {{ t('teams.noIssues') }}
      </div>
      <ul
        v-else
        class="space-y-2"
      >
        <li
          v-for="issue in issues"
          :key="issue.id"
          class="border rounded-md p-3 cursor-pointer hover:bg-accent/40"
          @click="openIssue(issue.id)"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="font-medium truncate">
              <span class="text-muted-foreground mr-2">#{{ issue.number }}</span>
              {{ issue.title }}
            </div>
            <Badge :variant="badgeVariantForStatus(issue.status)">
              {{ issue.status }}
            </Badge>
          </div>
          <div class="text-xs text-muted-foreground mt-1 truncate">
            {{ issue.description }}
          </div>
        </li>
      </ul>
    </div>

    <Dialog v-model:open="showCreate">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('teams.newIssue') }}</DialogTitle>
        </DialogHeader>
        <form
          class="space-y-3"
          @submit.prevent="submitIssue"
        >
          <div>
            <label class="text-sm">{{ t('teams.issueTitle') }}</label>
            <Input
              v-model="issueForm.title"
              required
            />
          </div>
          <div>
            <label class="text-sm">{{ t('teams.issueDescription') }}</label>
            <Textarea
              v-model="issueForm.description"
              rows="5"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              @click="showCreate = false"
            >
              {{ t('common.cancel') }}
            </Button>
            <Button type="submit">
              {{ t('common.save') }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import { Plus } from 'lucide-vue-next'
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Textarea,
} from '@memohai/ui'
import {
  getTeamsByTeamId,
  getTeamsByTeamIdIssues,
  postTeamsByTeamIdIssues,
} from '@memohai/sdk'
import { resolveApiErrorMessage } from '@/utils/api-error'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const queryCache = useQueryCache()

const teamId = computed(() => String(route.params.teamId ?? ''))

const { data: teamData } = useQuery({
  key: () => ['team', teamId.value],
  query: async () => {
    const { data, error } = await getTeamsByTeamId({ path: { team_id: teamId.value } })
    if (error) throw error
    return data
  },
  enabled: () => !!teamId.value,
})

const { data: issuesData } = useQuery({
  key: () => ['team', teamId.value, 'issues'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdIssues({ path: { team_id: teamId.value } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value,
})

const team = computed(() => teamData.value)
const issues = computed(() => issuesData.value ?? [])

const showCreate = ref(false)
const issueForm = reactive({
  title: '',
  description: '',
})

const { mutate: doCreate } = useMutation({
  mutation: async () => {
    const { data, error } = await postTeamsByTeamIdIssues({
      path: { team_id: teamId.value },
      body: {
        title: issueForm.title.trim(),
        description: issueForm.description.trim(),
      },
    })
    if (error) throw error
    return data
  },
  onSuccess: (issue) => {
    toast.success(t('teams.issueCreated'))
    showCreate.value = false
    issueForm.title = ''
    issueForm.description = ''
    void queryCache.invalidateQueries({ key: ['team', teamId.value, 'issues'] })
    if (issue?.id) {
      router.push({ name: 'team-issue', params: { teamId: teamId.value, issueId: issue.id } })
    }
  },
  onError: (err) => {
    toast.error(resolveApiErrorMessage(err, t('teams.issueCreateFailed')))
  },
})

function submitIssue() {
  if (!issueForm.title.trim()) return
  doCreate()
}

function openIssue(id?: string) {
  if (!id) return
  router.push({ name: 'team-issue', params: { teamId: teamId.value, issueId: id } })
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
