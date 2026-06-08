import { createSettingsClient, type SettingsClient } from "./settingsClient";
// @ts-expect-error binding details stay private to the API client boundary.
import type { SettingsBinding } from "./settingsClient";
import type { ApiResult } from "./result";
import type {
  ConfigError,
  SettingsView,
  UpdateSettingsInput,
} from "@/types/settings";

const settingsView: SettingsView = {
  appearance: {
    language: "zh",
    theme: "system",
  },
  paths: {
    dataDir: "E:/loomidbx/data",
    configFile: "E:/loomidbx/config.json",
  },
  development: {
    mode: "development",
    useIsolatedDataDir: true,
    diagnosticsEnabled: false,
  },
  integrations: {
    account: {
      status: "unavailable",
    },
    llm: {
      configured: false,
    },
  },
  privacy: {
    localOnly: true,
    telemetryEnabled: false,
  },
};

const updateInput: UpdateSettingsInput = {
  appearance: {
    theme: "dark",
  },
};

type TestSettingsBinding = NonNullable<
  Parameters<typeof createSettingsClient>[0]
>;

const successBinding: TestSettingsBinding = {
  GetSettings: () => settingsView,
  UpdateSettings: async () => settingsView,
};

const validationError: ConfigError = {
  code: "VALIDATION_FAILED",
  message: "设置校验失败",
  issues: [
    {
      path: "appearance.theme",
      code: "VALIDATION_FAILED",
      severity: "error",
      message: "主题值无效",
    },
  ],
};

const failingBinding: TestSettingsBinding = {
  GetSettings: () => {
    throw validationError;
  },
  UpdateSettings: () => Promise.reject(validationError),
};

const successClient: SettingsClient = createSettingsClient(successBinding);
const failingClient: SettingsClient = createSettingsClient(failingBinding);

const readResult: Promise<ApiResult<SettingsView>> =
  successClient.getSettings();
const updateResult: Promise<ApiResult<SettingsView>> =
  successClient.updateSettings(updateInput);
const failedReadResult: Promise<ApiResult<SettingsView>> =
  failingClient.getSettings();
const failedUpdateResult: Promise<ApiResult<SettingsView>> =
  failingClient.updateSettings(updateInput);

async function confirmSettingsClientBranches(
  client: SettingsClient,
): Promise<string> {
  const result = await client.getSettings();
  if (result.ok) {
    return result.data.appearance.theme;
  }

  const firstIssue = result.error.issues?.[0];
  return firstIssue?.path ?? result.error.message;
}

// @ts-expect-error 页面层不应直接接收 generated binding 形态。
const pageReadableClient: SettingsClient = successBinding;

export {
  confirmSettingsClientBranches,
  failedReadResult,
  failedUpdateResult,
  pageReadableClient,
  readResult,
  updateResult,
};
