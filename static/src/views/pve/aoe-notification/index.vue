<template>
  <div class="aoe-notification-page art-full-height">
    <ElCard class="toolbar-card" shadow="never">
      <div class="toolbar">
        <ElButton type="primary" :icon="Plus" @click="openCreateDialog">
          {{ t('aoeNotification.create') }}
        </ElButton>

        <ElButton :loading="loading" @click="loadNotifications">
          <el-icon class="mr-1"><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </ElButton>
      </div>
    </ElCard>

    <div v-loading="loading" class="notification-list">
      <ElCard
        v-for="notification in notifications"
        :key="notification.id"
        class="notification-card"
        shadow="hover"
      >
        <template #header>
          <div class="card-header">
            <span class="creator-name">{{ notification.creator_nickname }}</span>
          </div>
        </template>

        <div class="card-body">
          <div class="meta-row">
            <span class="meta-label">{{ t('common.createdAt') }}</span>
            <span class="meta-value">{{ notification.created_at }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">{{ t('aoeNotification.system') }}</span>
            <span class="meta-value">{{ notification.system }}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">{{ t('aoeNotification.type') }}</span>
            <ElTag size="small" type="warning">{{ notification.type }}</ElTag>
          </div>
          <div class="meta-row remark-row">
            <span class="meta-label">{{ t('aoeNotification.remark') }}</span>
            <span class="meta-value remark-value">{{ notification.remark || '-' }}</span>
          </div>
        </div>

        <template v-if="isOwnNotification(notification)" #footer>
          <div class="card-footer">
            <ElButton size="small" type="primary" link @click="openEditDialog(notification)">
              {{ t('common.edit') }}
            </ElButton>
            <ElButton size="small" type="danger" link @click="handleDelete(notification)">
              {{ t('common.delete') }}
            </ElButton>
          </div>
        </template>
      </ElCard>

      <ElEmpty v-if="notifications.length === 0" :description="t('aoeNotification.empty')" />
    </div>

    <ElDialog
      v-model="dialogVisible"
      :title="editingNotification ? t('aoeNotification.edit') : t('aoeNotification.create')"
      width="520px"
      destroy-on-close
    >
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="88px">
        <ElFormItem :label="t('aoeNotification.system')" prop="system">
          <ElInput
            v-model="formData.system"
            maxlength="128"
            show-word-limit
            :placeholder="t('aoeNotification.systemPlaceholder')"
          />
        </ElFormItem>
        <ElFormItem :label="t('aoeNotification.type')" prop="type">
          <ElInput
            v-model="formData.type"
            maxlength="128"
            show-word-limit
            :placeholder="t('aoeNotification.typePlaceholder')"
          />
        </ElFormItem>
        <ElFormItem :label="t('aoeNotification.remark')" prop="remark">
          <ElInput
            v-model="formData.remark"
            type="textarea"
            :rows="4"
            maxlength="1024"
            show-word-limit
            :placeholder="t('aoeNotification.remarkPlaceholder')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="saving" @click="handleSave">
          {{ t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    ElButton,
    ElCard,
    ElDialog,
    ElEmpty,
    ElForm,
    ElFormItem,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElTag,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { Plus, Refresh } from '@element-plus/icons-vue'
  import { useUserStore } from '@/store/modules/user'
  import {
    createAOENotification,
    deleteAOENotification,
    fetchAOENotificationList,
    updateAOENotification
  } from '@/api/pve'

  defineOptions({ name: 'AOENotification' })

  const { t } = useI18n()
  const userStore = useUserStore()
  const loading = ref(false)
  const saving = ref(false)
  const notifications = ref<Api.PVE.AOENotification[]>([])
  const dialogVisible = ref(false)
  const editingNotification = ref<Api.PVE.AOENotification | null>(null)
  const formRef = ref<FormInstance>()
  const formData = reactive<Api.PVE.AOENotificationRequest>({
    system: '',
    type: '',
    remark: ''
  })

  const formRules: FormRules = {
    system: [{ required: true, message: t('aoeNotification.systemRequired'), trigger: 'blur' }],
    type: [{ required: true, message: t('aoeNotification.typeRequired'), trigger: 'blur' }]
  }

  const currentUserId = computed(() => userStore.info?.userId ?? 0)

  const isOwnNotification = (notification: Api.PVE.AOENotification) => {
    return notification.creator_user_id === currentUserId.value
  }

  const loadNotifications = async () => {
    loading.value = true
    try {
      const result = await fetchAOENotificationList({ current: 1, size: 100 })
      notifications.value = result.list ?? []
    } finally {
      loading.value = false
    }
  }

  const resetForm = () => {
    formData.system = ''
    formData.type = ''
    formData.remark = ''
    formRef.value?.clearValidate()
  }

  const openCreateDialog = () => {
    editingNotification.value = null
    resetForm()
    dialogVisible.value = true
  }

  const openEditDialog = (notification: Api.PVE.AOENotification) => {
    editingNotification.value = notification
    formData.system = notification.system
    formData.type = notification.type
    formData.remark = notification.remark
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!formRef.value) return
    await formRef.value.validate()

    saving.value = true
    try {
      const payload = {
        system: formData.system.trim(),
        type: formData.type.trim(),
        remark: formData.remark.trim()
      }
      if (editingNotification.value) {
        await updateAOENotification(editingNotification.value.id, payload)
        ElMessage.success(t('aoeNotification.updateSuccess'))
      } else {
        await createAOENotification(payload)
        ElMessage.success(t('aoeNotification.createSuccess'))
      }
      dialogVisible.value = false
      await loadNotifications()
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (notification: Api.PVE.AOENotification) => {
    try {
      await ElMessageBox.confirm(
        t('aoeNotification.deleteConfirm', { system: notification.system }),
        t('common.tips'),
        {
          type: 'warning',
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel')
        }
      )
    } catch {
      return
    }
    await deleteAOENotification(notification.id)
    ElMessage.success(t('aoeNotification.deleteSuccess'))
    await loadNotifications()
  }

  onMounted(loadNotifications)
</script>

<style scoped lang="scss">
  .aoe-notification-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .toolbar-card {
    flex-shrink: 0;
  }

  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .notification-list {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    align-content: flex-start;
    gap: 16px;
    min-height: 0;
  }

  .notification-card {
    flex: 1 1 280px;
    max-width: 420px;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .creator-name {
    font-size: 15px;
    font-weight: 600;
  }

  .card-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .meta-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .meta-label {
    color: var(--art-text-gray-600);
    font-size: 13px;
  }

  .meta-value {
    color: var(--art-text-gray-900);
    font-size: 14px;
    font-weight: 500;
    text-align: right;
  }

  .remark-row {
    align-items: flex-start;
  }

  .remark-value {
    max-width: 260px;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .card-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  :deep(.el-empty) {
    width: 100%;
    min-height: 260px;
  }

  @media (max-width: 640px) {
    .toolbar {
      align-items: stretch;
      flex-direction: column;
    }

    .notification-card {
      max-width: none;
    }
  }
</style>
