import {
  GetSettings as generatedGetSettings,
  UpdateSettings as generatedUpdateSettings,
} from '../../generated/settings'
import type { SettingsView, UpdateSettingsInput } from '@/types/settings'
import type { ApiResult } from './result'
import { err, ok } from './result'

interface SettingsBinding {
  GetSettings(): Promise<SettingsView> | SettingsView
  UpdateSettings(
    input: UpdateSettingsInput,
  ): Promise<SettingsView> | SettingsView
}

export interface SettingsClient {
  getSettings(): Promise<ApiResult<SettingsView>>
  updateSettings(input: UpdateSettingsInput): Promise<ApiResult<SettingsView>>
}

export function createSettingsClient(
  binding: SettingsBinding = {
    GetSettings: generatedGetSettings,
    UpdateSettings: generatedUpdateSettings,
  },
): SettingsClient {
  return {
    async getSettings(): Promise<ApiResult<SettingsView>> {
      try {
        return ok(await binding.GetSettings())
      } catch (error) {
        return err(error)
      }
    },

    async updateSettings(
      input: UpdateSettingsInput,
    ): Promise<ApiResult<SettingsView>> {
      try {
        return ok(await binding.UpdateSettings(input))
      } catch (error) {
        return err(error)
      }
    },
  }
}

export const settingsClient = createSettingsClient()
