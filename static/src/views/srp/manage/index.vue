<!-- SRP 补损审批管理页面 -->
<template>
  <div class="srp-manage-page art-full-height">
    <ElCard class="art-search-card" shadow="never">
      <div class="flex items-center gap-3 flex-wrap">
        <ElSelect
          v-model="filter.review_status"
          :placeholder="$t('srp.apply.columns.reviewStatus')"
          clearable
          style="width: 130px"
          @change="handleSearch"
        >
          <ElOption :label="$t('srp.status.pending')" value="pending" />
          <ElOption :label="$t('srp.status.approved')" value="approved" />
          <ElOption :label="$t('srp.status.rejected')" value="rejected" />
        </ElSelect>
        <ElSelect
          v-model="filter.payout_status"
          :placeholder="$t('srp.apply.columns.payoutStatus')"
          clearable
          style="width: 130px"
          @change="handleSearch"
        >
          <ElOption :label="$t('srp.status.unpaid')" value="pending" />
          <ElOption :label="$t('srp.status.paid')" value="paid" />
        </ElSelect>
        <ElSelect
          v-model="filter.fleet_id"
          :placeholder="$t('srp.manage.selectFleet')"
          clearable
          filterable
          style="width: 220px"
          @change="handleSearch"
        >
          <ElOption v-for="f in fleets" :key="f.id" :label="f.title" :value="f.id" />
        </ElSelect>
        <ElButton type="primary" @click="handleSearch">{{ $t('srp.manage.searchBtn') }}</ElButton>
        <ElButton @click="resetFilter">{{ $t('srp.manage.resetBtn') }}</ElButton>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDialog
      v-model="reviewDialogVisible"
      :title="
        reviewAction === 'approve' ? $t('srp.manage.approveDialog') : $t('srp.manage.rejectDialog')
      "
      width="460px"
    >
      <ElForm label-width="90px">
        <ElFormItem :label="$t('srp.manage.finalAmount')" v-if="reviewAction === 'approve'">
          <ElInputNumber
            v-model="reviewForm.final_amount"
            :min="0"
            :precision="2"
            :step="1000000"
            style="width: 100%"
          />
          <div class="text-xs text-gray-400 mt-1">{{ $t('srp.manage.finalAmountHint') }}</div>
        </ElFormItem>
        <ElFormItem :label="$t('srp.manage.reviewNote')" :required="reviewAction === 'reject'">
          <ElInput
            v-model="reviewForm.review_note"
            type="textarea"
            :rows="3"
            :placeholder="
              reviewAction === 'reject'
                ? $t('srp.manage.rejectNotePlaceholder')
                : $t('srp.manage.optionalPlaceholder')
            "
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="reviewDialogVisible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
        <ElButton
          :type="reviewAction === 'approve' ? 'success' : 'danger'"
          :loading="actionLoading"
          @click="handleReview"
        >
          {{
            reviewAction === 'approve'
              ? $t('srp.manage.confirmApprove')
              : $t('srp.manage.confirmReject')
          }}
        </ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="payoutDialogVisible" :title="$t('srp.manage.payoutDialog')" width="420px">
      <div class="mb-2"
        >{{ $t('srp.manage.payoutCharacter')
        }}<strong>{{ payoutTarget?.character_name }}</strong></div
      >
      <div class="mb-4"
        >{{ $t('srp.manage.payoutAmount')
        }}<strong class="text-blue-600"
          >{{ formatISK(payoutTarget?.final_amount ?? 0) }} ISK</strong
        ></div
      >
      <ElFormItem :label="$t('srp.manage.overrideAmount')">
        <ElInputNumber
          v-model="payoutOverrideAmount"
          :min="0"
          :precision="2"
          :step="1000000"
          style="width: 100%"
        />
        <div class="text-xs text-gray-400 mt-1">{{ $t('srp.manage.overrideAmountHint') }}</div>
      </ElFormItem>
      <template #footer>
        <ElButton @click="payoutDialogVisible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
        <ElButton type="primary" :loading="actionLoading" @click="handlePayout">{{
          $t('srp.manage.confirmPayout')
        }}</ElButton>
      </template>
    </ElDialog>

    <!-- KM 预览弹窗 -->
    <KmPreviewDialog v-model="kmPreviewVisible" :killmail-id="previewKillmailId" />
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import {
    ElCard,
    ElTag,
    ElButton,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElInput,
    ElLink,
    ElMessage
  } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import KmPreviewDialog from '@/components/business/KmPreviewDialog.vue'
  import { fetchFleetList } from '@/api/fleet'
  import { fetchApplicationList, reviewApplication, payoutApplication } from '@/api/srp'
  import { useNameResolver } from '@/hooks'

  defineOptions({ name: 'SrpManage' })

  const { t } = useI18n()
  const { getName, resolve: resolveNames } = useNameResolver()

  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const loadFleets = async () => {
    try {
      const res = await fetchFleetList({ size: 200 } as any)
      fleets.value = res?.records ?? []
    } catch {
      fleets.value = []
    }
  }

  const filter = reactive({ review_status: '', payout_status: '', fleet_id: '' })

  type SrpApp = Api.Srp.Application
  type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

  const reviewStatusType = (s: string): TagType =>
    (({ pending: 'info', approved: 'success', rejected: 'danger' }) as Record<string, TagType>)[
      s
    ] ?? 'info'
  const reviewStatusLabel = (s: string) =>
    ({
      pending: t('srp.status.pending'),
      approved: t('srp.status.approved'),
      rejected: t('srp.status.rejected')
    })[s as 'pending' | 'approved' | 'rejected'] ?? s
  const payoutStatusType = (s: string): TagType => (s === 'paid' ? 'success' : 'warning')

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData,
    searchParams
  } = useTable({
    core: {
      apiFn: fetchApplicationList,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'character_name',
          label: t('srp.manage.columns.character'),
          width: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'ship_type_id',
          label: t('srp.manage.columns.ship'),
          width: 180,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', {}, getName(row.ship_type_id, `TypeID: ${row.ship_type_id}`))
        },
        {
          prop: 'solar_system_id',
          label: t('srp.manage.columns.system'),
          width: 140,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', {}, getName(row.solar_system_id, String(row.solar_system_id)))
        },
        {
          prop: 'killmail_id',
          label: t('srp.manage.columns.killId'),
          width: 110,
          formatter: (row: SrpApp) =>
            h(
              ElLink,
              {
                href: `https://zkillboard.com/kill/${row.killmail_id}/`,
                target: '_blank',
                type: 'primary'
              },
              () => String(row.killmail_id)
            )
        },
        {
          prop: 'killmail_time',
          label: t('srp.manage.columns.kmTime'),
          width: 175,
          formatter: (row: SrpApp) => h('span', {}, formatTime(row.killmail_time))
        },
        {
          prop: 'corporation_id',
          label: t('srp.manage.columns.corporation'),
          width: 180,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              {},
              getName(row.corporation_id, row.corporation_id ? `ID: ${row.corporation_id}` : '-')
            )
        },
        {
          prop: 'alliance_id',
          label: t('srp.manage.columns.alliance'),
          width: 180,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              {},
              getName(row.alliance_id, row.alliance_id ? `ID: ${row.alliance_id}` : '-')
            )
        },
        {
          prop: 'recommended_amount',
          label: t('srp.manage.columns.recommendedAmount'),
          width: 140,
          formatter: (row: SrpApp) => h('span', {}, formatISK(row.recommended_amount))
        },
        {
          prop: 'final_amount',
          label: t('srp.manage.columns.finalAmount'),
          width: 140,
          formatter: (row: SrpApp) =>
            h('span', { class: 'font-semibold text-blue-600' }, formatISK(row.final_amount))
        },
        {
          prop: 'review_status',
          label: t('srp.manage.columns.review'),
          width: 100,
          formatter: (row: SrpApp) =>
            h(ElTag, { type: reviewStatusType(row.review_status), size: 'small' }, () =>
              reviewStatusLabel(row.review_status)
            )
        },
        {
          prop: 'payout_status',
          label: t('srp.manage.columns.payout'),
          width: 100,
          formatter: (row: SrpApp) =>
            h(ElTag, { type: payoutStatusType(row.payout_status), size: 'small' }, () =>
              row.payout_status === 'paid' ? t('srp.status.paid') : t('srp.status.unpaid')
            )
        },
        {
          prop: 'actions',
          label: t('srp.manage.columns.action'),
          width: 220,
          fixed: 'right',
          formatter: (row: SrpApp) => {
            const btns: ReturnType<typeof h>[] = [
              h(ArtButtonTable, { type: 'view', onClick: () => openKmPreview(row) })
            ]
            if (row.review_status === 'pending') {
              btns.push(
                h(
                  ElButton,
                  {
                    size: 'small',
                    type: 'success',
                    onClick: () => openReviewDialog(row, 'approve')
                  },
                  () => t('srp.manage.approveBtn')
                ),
                h(
                  ElButton,
                  {
                    size: 'small',
                    type: 'danger',
                    onClick: () => openReviewDialog(row, 'reject')
                  },
                  () => t('srp.manage.rejectBtn')
                )
              )
            } else if (row.review_status === 'approved' && row.payout_status === 'pending') {
              btns.push(
                h(
                  ElButton,
                  { size: 'small', type: 'primary', onClick: () => openPayoutDialog(row) },
                  () => t('srp.manage.payoutBtn')
                )
              )
            }
            return h('div', { class: 'flex gap-1' }, btns)
          }
        }
      ]
    }
  })

  watch(data, async (list) => {
    if (list.length) await resolveManageNames(list)
  })

  /** 收集申请列表中所有需要解析的 ID，一次性查询 */
  const resolveManageNames = async (list: Api.Srp.Application[]) => {
    const typeIds = new Set<number>()
    const solarIds = new Set<number>()
    const esiIds = new Set<number>()
    for (const app of list) {
      if (app.ship_type_id) typeIds.add(app.ship_type_id)
      if (app.solar_system_id) solarIds.add(app.solar_system_id)
      if (app.corporation_id) esiIds.add(app.corporation_id)
      if (app.alliance_id) esiIds.add(app.alliance_id)
    }
    await resolveNames({
      ids: {
        ...(typeIds.size ? { type: [...typeIds] } : {}),
        ...(solarIds.size ? { solar_system: [...solarIds] } : {})
      },
      esi: esiIds.size ? [...esiIds] : undefined
    })
  }

  const handleSearch = () => {
    Object.assign(searchParams, {
      review_status: filter.review_status || undefined,
      payout_status: filter.payout_status || undefined,
      fleet_id: filter.fleet_id || undefined
    })
    getData()
  }
  const resetFilter = () => {
    filter.review_status = ''
    filter.payout_status = ''
    filter.fleet_id = ''
    Object.assign(searchParams, {
      review_status: undefined,
      payout_status: undefined,
      fleet_id: undefined
    })
    getData()
  }

  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'approve' | 'reject'>('approve')
  const reviewTarget = ref<Api.Srp.Application | null>(null)
  const reviewForm = reactive({ review_note: '', final_amount: 0 })
  const actionLoading = ref(false)

  const openReviewDialog = (row: Api.Srp.Application, action: 'approve' | 'reject') => {
    reviewTarget.value = row
    reviewAction.value = action
    reviewForm.review_note = ''
    reviewForm.final_amount = 0
    reviewDialogVisible.value = true
  }

  const handleReview = async () => {
    if (!reviewTarget.value) return
    if (reviewAction.value === 'reject' && !reviewForm.review_note) {
      ElMessage.warning(t('srp.manage.rejectRequired'))
      return
    }
    actionLoading.value = true
    try {
      await reviewApplication(reviewTarget.value.id, {
        action: reviewAction.value,
        review_note: reviewForm.review_note,
        final_amount: reviewForm.final_amount
      })
      ElMessage.success(
        reviewAction.value === 'approve'
          ? t('srp.manage.approveSuccess')
          : t('srp.manage.rejectSuccess')
      )
      reviewDialogVisible.value = false
      refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  const payoutDialogVisible = ref(false)
  const payoutTarget = ref<Api.Srp.Application | null>(null)
  const payoutOverrideAmount = ref(0)

  const openPayoutDialog = (row: Api.Srp.Application) => {
    payoutTarget.value = row
    payoutOverrideAmount.value = 0
    payoutDialogVisible.value = true
  }

  const handlePayout = async () => {
    if (!payoutTarget.value) return
    actionLoading.value = true
    try {
      await payoutApplication(payoutTarget.value.id, { final_amount: payoutOverrideAmount.value })
      ElMessage.success(t('srp.manage.payoutSuccess'))
      payoutDialogVisible.value = false
      refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(
      v ?? 0
    )

  /* ── KM 预览 ── */
  const kmPreviewVisible = ref(false)
  const previewKillmailId = ref(0)
  const openKmPreview = (row: Api.Srp.Application) => {
    previewKillmailId.value = row.killmail_id
    kmPreviewVisible.value = true
  }

  onMounted(() => {
    loadFleets()
  })
</script>

<style scoped></style>
