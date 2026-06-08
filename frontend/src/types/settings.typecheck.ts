import type {
  ConfigIssue,
  ConfigIssueCode,
  ConfigIssueSeverity,
  SettingsView,
  UpdateSettingsInput,
} from './settings'

const issue: ConfigIssue = {
  path: 'appearance.theme',
  code: 'VALIDATION_FAILED',
  severity: 'error',
  message: '主题值无效',
}

const issueCode: ConfigIssueCode = issue.code
const issueSeverity: ConfigIssueSeverity = issue.severity

const view: SettingsView = {
  appearance: {
    language: 'zh',
    theme: 'system',
  },
  paths: {
    dataDir: 'E:/loomidbx/data',
    configFile: 'E:/loomidbx/config.json',
  },
  development: {
    mode: 'development',
    useIsolatedDataDir: true,
    diagnosticsEnabled: false,
  },
  integrations: {
    account: {
      status: 'unavailable',
    },
    llm: {
      configured: false,
    },
  },
  privacy: {
    localOnly: true,
    telemetryEnabled: false,
  },
}

const update: UpdateSettingsInput = {
  appearance: {
    theme: 'dark',
  },
  paths: {
    dataDir: 'E:/loomidbx/data',
  },
  development: {
    mode: 'test',
  },
  privacy: {
    telemetryEnabled: false,
  },
}

// @ts-expect-error 明文凭据不属于设置 DTO 边界。
type LLMPlaintextCredential = SettingsView['integrations']['llm']['credential']

// @ts-expect-error 更新输入不能接受完整配置文件路径。
type ConfigFileUpdate = NonNullable<UpdateSettingsInput['paths']>['configFile']

export { issueCode, issueSeverity, update, view }
export type { ConfigFileUpdate, LLMPlaintextCredential }
