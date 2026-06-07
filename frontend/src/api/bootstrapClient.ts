import { BootstrapStatus as generatedBootstrapStatus } from '../../generated/bootstrap'
import type { BootstrapStatus } from '@/types/bootstrap'
import type { ApiResult } from './result'
import { err, ok } from './result'

export interface BootstrapBinding {
  BootstrapStatus(): Promise<BootstrapStatus> | BootstrapStatus
}

export function createBootstrapClient(binding: BootstrapBinding = { BootstrapStatus: generatedBootstrapStatus }) {
  return {
    async getStatus(): Promise<ApiResult<BootstrapStatus>> {
      try {
        return ok(await binding.BootstrapStatus())
      } catch (error) {
        return err(error)
      }
    },
  }
}

export const bootstrapClient = createBootstrapClient()
