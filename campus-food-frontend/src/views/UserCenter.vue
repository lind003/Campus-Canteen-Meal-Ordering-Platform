<template>
  <div class="container">
      <div class="d-flex justify-content-between align-items-center mb-3">
    <h4 class="mb-0">个人中心</h4>
    <button class="btn btn-sm btn-outline-secondary" @click="goHome">
      返回主页
    </button>
  </div>
    <div class="card mb-4">
      <div class="card-body">
        <h5>基本信息</h5>
        <div>姓名：{{ profile.name ||'未填写'}}</div>
        <div>手机：{{ profile.phone }}</div>
        <div>学院：{{ profile.college }}</div>
        <div>年级：{{ profile.grade }}</div>
        <div>校园卡号：{{ profile.card_no||'未填写' }}</div>
        <div>角色：{{ profile.role==='runner'?'跑腿员':'普通用户' }}</div>
        <button class="btn btn-sm btn-outline-primary mt-2" @click="openEditModal">编辑资料</button>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
        <h5>账号安全</h5>
        <button class="btn btn-sm btn-outline-warning" @click="openPwd=true">修改密码</button>
      </div>
    </div>

    <!-- 编辑资料弹窗 -->
    <div v-if="openEdit" class="modal-overlay" @click.self="openEdit=false">
      <div class="modal-content">
        <h5>编辑资料</h5>
        <form @submit.prevent="submitEdit">
          <div class="form-group"><label>姓名</label><input v-model="form.name" required></div>
          <div class="form-group"><label>手机</label><input v-model="form.phone" required></div>
          <div class="form-group"><label>学院</label><input v-model="form.college"></div>
          <div class="form-group"><label>年级</label><input v-model="form.grade" ></div>
          <div class="form-group"><label>校园卡号</label><input v-model="form.card_no"></div>
          <div class="form-group"><label>当前密码</label><input v-model="form.password" type="password" required></div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary">保存</button>
            <button type="button" class="btn btn-secondary" @click="openEdit=false">取消</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 修改密码弹窗 -->
    <div v-if="openPwd" class="modal-overlay" @click.self="openPwd=false">
      <div class="modal-content">
        <h5>修改密码</h5>
        <form @submit.prevent="submitPwd">
          <div class="form-group"><label>原密码</label><input v-model="pwdForm.old_password" type="password" required></div>
          <div class="form-group"><label>新密码</label><input v-model="pwdForm.new_password" type="password" required></div>
          <div class="form-actions">
            <button type="submit" class="btn btn-warning">确认修改</button>
            <button type="button" class="btn btn-secondary" @click="openPwd=false">取消</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/utils/api'
import { useRouter } from 'vue-router'

const auth   = useAuthStore()
const router = useRouter()
const profile = ref({
  name: '',
  phone: '',
  college: '',
  grade: '',
  card_no: '',
  role: ''
})

const openEdit = ref(false)
const openPwd  = ref(false)
const form = ref({
  name: '',
  phone: '',
  college: '',
  grade: '',
  card_no: '',
  password: ''
})
const pwdForm = reactive({ old_password:'', new_password:'' })

//返回主页
function goHome(){
    router.push('/')
}
//编辑资料
function openEditModal(){
  form.value.name = profile.value.name
  form.value.phone = profile.value.phone
  form.value.college = profile.value.college
  form.value.grade = profile.value.grade
  form.value.card_no = profile.value.card_no
  form.value.password = ''
  openEdit.value = true

}
onMounted(async () => {
  console.log(' UserCenter 组件已挂载')
  try {
    const { data :res} = await authAPI.profile()
    console.log('获取个人信息成功:', res)
    Object.assign(profile.value, res.data)
  } catch (e) {
    console.error(' 加载个人信息失败:', e)
    alert('加载个人信息失败')
  }
})

async function submitEdit(){
  try {
    // 只提交普通对象
    const payload = {
      name: form.value.name,
      phone: form.value.phone,
      college: form.value.college,
      grade: form.value.grade,
      card_no: form.value.card_no,
      password: form.value.password
    }

    await authAPI.updateProfile(payload)
    await reload()
    openEdit.value = false
    alert('资料已更新')
  } catch (e) {
    alert('保存失败：' + (e.response?.data?.message || e.message))
  }
}
async function submitPwd(){
  await authAPI.updatePassword(pwdForm)
  alert('密码已修改，请重新登录')
  auth.logout()
  router.push('/')
}
async function reload(){
  const { data:res } = await authAPI.profile()
  Object.assign(profile.value, res.data)
}
</script>

<style scoped>
.modal-overlay{ position:fixed; inset:0; background:rgba(0,0,0,.5); display:flex; align-items:center; justify-content:center; z-index:999; }
.modal-content{ background:#fff; padding:2rem; border-radius:8px; width:90%; max-width:400px; }
.form-group{ margin-bottom:1rem; }
.form-actions{ display:flex; gap:1rem; margin-top:1.5rem; }
</style>