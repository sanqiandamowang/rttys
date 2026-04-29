<template>
  <el-container>
    <el-header class="header">
      <el-space>
        <el-button type="primary" round icon="Refresh" @click="handleRefresh" :disabled="loading">{{ $t('Refresh List')
        }}</el-button>
        <el-button @click="goBack">{{ $t('Back') }}</el-button>
        <el-button type="danger" icon="Delete" @click="handleDeleteAll" :disabled="!selectedHistories.length">{{
          $t('Delete Selected') }}</el-button>
      </el-space>
    </el-header>
    <el-main>
      <el-card>
        <el-table height="calc(100vh - 225px)" :loading="loading" :data="histories"
          :empty-text="$t('No history records')">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="device_id" :label="$t('Device ID')" width="200" />
          <el-table-column prop="group_name" :label="$t('Group')" width="150" />
          <el-table-column prop="ip_addr" :label="$t('IP Address')" width="150" />
          <el-table-column prop="description" :label="$t('Description')" show-overflow-tooltip width="200" />
          <el-table-column :label="$t('Online Time')" width="180">
            <template #default="{ row }">
              <span>{{ formatTime(row.online_time) }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('Offline Time')" width="180">
            <template #default="{ row }">
              <span>{{ row.offline_time ? formatTime(row.offline_time) : $t('Online') }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('Duration')" width="100">
            <template #default="{ row }">
              <span>{{ row.duration ? formatDuration(row.duration) : '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="$t('Actions')" width="100">
            <template #default="{ row }">
              <el-button type="danger" size="small" icon="Delete" @click="handleDelete(row.id)">{{ $t('Delete')
              }}</el-button>
            </template>
          </el-table-column>
        </el-table>
        <template #footer>
          <el-pagination background layout="prev, pager, next, total,sizes" :total="histories.length"
            @change="handlePageChange" class="pagination" />
        </template>
      </el-card>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

const loading = ref(true)
const histories = ref([])
const selectedHistories = ref([])
const currentPage = ref(1)
const pageSize = ref(10)

const deviceId = computed(() => route.params.devid || '')
const group = computed(() => route.query.group || '')

const formatTime = (t) => {
  if (!t) return ''
  const date = new Date(t)
  return date.toLocaleString()
}

const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

const getHistory = () => {
  loading.value = true
  let url = deviceId.value ? `/api/history/${deviceId.value}` : '/api/history'
  if (group.value && !deviceId.value) {
    url += `?group=${group.value}`
  }

  console.log('Fetching history from:', url, 'deviceId:', deviceId.value, 'group:', group.value)
  axios.get(url).then(res => {
    console.log('History response:', res.data)
    loading.value = false
    histories.value = res.data
  }).catch(err => {
    loading.value = false
    console.error('Failed to fetch history:', err)
  })
}

const handleRefresh = () => {
  loading.value = true
  setTimeout(getHistory, 500)
}

const handlePageChange = (page, size) => {
  currentPage.value = page
  pageSize.value = size
}

const handleDelete = (id) => {
  axios.post('/api/history/delete', { ids: [id] }).then(() => {
    getHistory()
    ElMessage.success('删除成功')
  }).catch(err => {
    console.error(err)
    ElMessage.error('删除失败')
  })
}

const handleDeleteAll = () => {
  if (!selectedHistories.value.length) return

  axios.post('/api/history/delete', { ids: selectedHistories.value.map(h => h.id) }).then(() => {
    getHistory()
    selectedHistories.value = []
    ElMessage.success('删除成功')
  }).catch(err => {
    console.error(err)
    ElMessage.error('删除失败')
  })
}

const goBack = () => {
  router.back()
}

onMounted(() => {
  getHistory()
})
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
}

.pagination {
  display: flex;
  justify-content: center;
}
</style>
