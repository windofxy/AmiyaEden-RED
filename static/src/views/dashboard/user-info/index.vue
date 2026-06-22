<template>
  <div class="dashboard-user-info-page art-full-height">
    <ElCard class="user-info-card" shadow="never">
      <template #header>
        <div class="card-title">
          <h2>{{ t('dashboardUserInfo.title') }}</h2>
          <p>{{ t('dashboardUserInfo.description') }}</p>
        </div>
      </template>

      <ElForm label-position="top" class="user-info-form">
        <ElFormItem :label="t('dashboardUserInfo.nickname')">
          <ElInput v-model="nickname" :readonly="!editing" maxlength="128" show-word-limit>
            <template #append>
              <ElButton :loading="saving" @click="handleNicknameAction">
                {{ editing ? t('common.save') : t('common.edit') }}
              </ElButton>
            </template>
          </ElInput>
          <p class="form-tip">{{ t('dashboardUserInfo.nicknameHint') }}</p>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElButton, ElCard, ElForm, ElFormItem, ElInput, ElMessage } from 'element-plus'
  import { fetchGetUserInfo, updateMyNickname } from '@/api/auth'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'DashboardUserInfo' })

  const { t } = useI18n()
  const userStore = useUserStore()
  const editing = ref(false)
  const saving = ref(false)
  const nickname = ref('')

  watch(
    () => userStore.getUserInfo.userName,
    (value) => {
      if (!editing.value) {
        nickname.value = value ?? ''
      }
    },
    { immediate: true }
  )

  const handleNicknameAction = async () => {
    if (!editing.value) {
      editing.value = true
      return
    }

    const nextNickname = nickname.value.trim()
    if (!nextNickname) {
      ElMessage.warning(t('dashboardUserInfo.nicknameRequired'))
      return
    }

    saving.value = true
    try {
      await updateMyNickname({ nickname: nextNickname })
      const userInfo = await fetchGetUserInfo()
      userStore.setUserInfo(userInfo)
      nickname.value = userInfo.userName
      editing.value = false
      ElMessage.success(t('dashboardUserInfo.saveSuccess'))
    } catch (error) {
      console.error(error)
      ElMessage.error(t('dashboardUserInfo.saveFailed'))
    } finally {
      saving.value = false
    }
  }
</script>

<style scoped lang="scss">
  .dashboard-user-info-page {
    display: flex;
    flex-direction: column;
  }

  .user-info-card {
    max-width: 720px;
  }

  .card-title {
    h2 {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
    }

    p {
      margin: 6px 0 0;
      color: var(--art-text-gray-600);
      font-size: 13px;
    }
  }

  .user-info-form {
    max-width: 520px;
  }

  .form-tip {
    margin: 8px 0 0;
    color: var(--art-text-gray-500);
    font-size: 12px;
    line-height: 1.5;
  }
</style>
