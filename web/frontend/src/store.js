import { reactive } from 'vue'

export const store = reactive({
  user: null,
  setupDone: null,
  setupStep: 0,
  centerName: '',
})
