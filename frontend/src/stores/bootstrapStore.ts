import { reactive } from 'vue'

type BootstrapState = 'loading' | 'ready' | 'error'

export interface BootstrapViewState {
  state: BootstrapState
  title: string
  message: string
}

const bootstrapState = reactive<BootstrapViewState>({
  state: 'ready',
  title: 'Bootstrap skeleton ready',
  message: '当前页面只验证工程骨架，完整业务工作流由后续 spec 实现。',
})

export function useBootstrapStore(): BootstrapViewState {
  return bootstrapState
}
