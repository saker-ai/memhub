<template>
  <section class="px-4 pt-2 pb-10 lg:px-6 md:pt-4 md:pb-12 space-y-6">
    <div class="flex items-center gap-3">
      <Button
        variant="ghost"
        size="sm"
        @click="router.push({ name: 'teams' })"
      >
        <ArrowLeft class="size-4" />
      </Button>
      <h1 class="text-lg font-semibold">
        {{ team?.name ?? t('teams.title') }}
      </h1>
    </div>

    <Card v-if="team">
      <CardHeader>
        <CardTitle>{{ t('teams.settings') }}</CardTitle>
      </CardHeader>
      <CardContent class="space-y-2 text-sm">
        <div><span class="text-muted-foreground">{{ t('teams.id') }}:</span> <code>{{ team.id }}</code></div>
        <div v-if="team.description">
          <span class="text-muted-foreground">{{ t('teams.description') }}:</span> {{ team.description }}
        </div>
        <div v-if="team.shared_dir_name">
          <span class="text-muted-foreground">{{ t('teams.sharedDir') }}:</span> <code>/team/{{ team.shared_dir_name }}</code>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="flex flex-row items-center justify-between">
        <CardTitle>{{ t('teams.members') }}</CardTitle>
        <Button
          size="sm"
          @click="showAdd = true"
        >
          <Plus class="size-4 mr-1" /> {{ t('teams.addMember') }}
        </Button>
      </CardHeader>
      <CardContent>
        <div
          v-if="members.length === 0"
          class="text-sm text-muted-foreground"
        >
          {{ t('teams.noMembers') }}
        </div>
        <ul
          v-else
          class="space-y-2"
        >
          <li
            v-for="m in members"
            :key="m.id"
            class="flex items-center justify-between text-sm"
          >
            <div>
              <span class="font-medium">{{ m.display_name || m.bot_id || m.user_id }}</span>
              <span class="ml-2 text-muted-foreground">{{ m.member_type }}</span>
              <span
                v-if="m.role"
                class="ml-2 text-muted-foreground"
              >· {{ m.role }}</span>
            </div>
            <Button
              variant="ghost"
              size="sm"
              @click="removeMember(m.id)"
            >
              {{ t('common.remove') }}
            </Button>
          </li>
        </ul>
      </CardContent>
    </Card>

    <Dialog v-model:open="showAdd">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ t('teams.addMember') }}</DialogTitle>
        </DialogHeader>
        <form
          class="space-y-3"
          @submit.prevent="submitMember"
        >
          <div>
            <label class="text-sm">{{ t('teams.memberType') }}</label>
            <select
              v-model="memberForm.member_type"
              class="w-full rounded border px-2 py-1 text-sm"
            >
              <option value="bot">
                bot
              </option>
              <option value="user">
                user
              </option>
            </select>
          </div>
          <div v-if="memberForm.member_type === 'bot'">
            <label class="text-sm">{{ t('teams.botId') }}</label>
            <BotSelect v-model="memberForm.bot_id" />
          </div>
          <div v-else>
            <label class="text-sm">{{ t('teams.userId') }}</label>
            <Input v-model="memberForm.user_id" />
          </div>
          <div>
            <label class="text-sm">{{ t('teams.role') }}</label>
            <Input
              v-model="memberForm.role"
              :placeholder="t('teams.rolePlaceholder')"
            />
          </div>
          <div>
            <label class="text-sm">{{ t('teams.memberInstructions') }}</label>
            <Textarea
              v-model="memberForm.instructions"
              rows="3"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              @click="showAdd = false"
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
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import { toast } from 'vue-sonner'
import { ArrowLeft, Plus } from 'lucide-vue-next'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Textarea,
} from '@memohai/ui'
import {
  deleteTeamsByTeamIdMembersByMemberId,
  getTeamsByTeamId,
  getTeamsByTeamIdMembers,
  postTeamsByTeamIdMembers,
} from '@memohai/sdk'
import BotSelect from '@/components/bot-select/index.vue'
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

const { data: membersData } = useQuery({
  key: () => ['team', teamId.value, 'members'],
  query: async () => {
    const { data, error } = await getTeamsByTeamIdMembers({ path: { team_id: teamId.value } })
    if (error) throw error
    return data ?? []
  },
  enabled: () => !!teamId.value,
})

const team = computed(() => teamData.value)
const members = computed(() => membersData.value ?? [])

const showAdd = ref(false)
const memberForm = reactive({
  member_type: 'bot' as 'bot' | 'user',
  bot_id: '',
  user_id: '',
  role: '',
  instructions: '',
})

const { mutate: doAdd } = useMutation({
  mutation: async () => {
    const { data, error } = await postTeamsByTeamIdMembers({
      path: { team_id: teamId.value },
      body: {
        member_type: memberForm.member_type,
        bot_id: memberForm.bot_id,
        user_id: memberForm.user_id,
        role: memberForm.role,
        instructions: memberForm.instructions,
      },
    })
    if (error) throw error
    return data
  },
  onSuccess: () => {
    toast.success(t('teams.memberAdded'))
    showAdd.value = false
    void queryCache.invalidateQueries({ key: ['team', teamId.value, 'members'] })
  },
  onError: (err) => {
    toast.error(resolveApiErrorMessage(err, t('teams.memberAddFailed')))
  },
})

function submitMember() {
  doAdd()
}

const { mutate: doRemove } = useMutation({
  mutation: async (memberId: string) => {
    const { error } = await deleteTeamsByTeamIdMembersByMemberId({
      path: { team_id: teamId.value, member_id: memberId },
    })
    if (error) throw error
  },
  onSuccess: () => {
    void queryCache.invalidateQueries({ key: ['team', teamId.value, 'members'] })
  },
  onError: (err) => {
    toast.error(resolveApiErrorMessage(err, t('teams.memberRemoveFailed')))
  },
})

function removeMember(id: string | undefined) {
  if (!id) return
  doRemove(id)
}
</script>
