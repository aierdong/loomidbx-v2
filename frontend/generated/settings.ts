import type {
  SettingsView as SettingsViewDto,
  UpdateSettingsInput as UpdateSettingsInputDto,
} from "@/types/settings";

const defaultSettingsView: SettingsViewDto = {
  appearance: {
    language: "zh",
    theme: "system",
  },
  paths: {
    dataDir: "",
    configFile: "",
  },
  development: {
    mode: "desktop",
    useIsolatedDataDir: false,
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

export async function GetSettings(): Promise<SettingsViewDto> {
  return defaultSettingsView;
}

export async function UpdateSettings(
  input: UpdateSettingsInputDto,
): Promise<SettingsViewDto> {
  return {
    ...defaultSettingsView,
    appearance: {
      ...defaultSettingsView.appearance,
      ...input.appearance,
    },
    paths: {
      ...defaultSettingsView.paths,
      dataDir: input.paths?.dataDir ?? defaultSettingsView.paths.dataDir,
    },
    development: {
      ...defaultSettingsView.development,
      ...input.development,
    },
    privacy: {
      ...defaultSettingsView.privacy,
      ...input.privacy,
    },
  };
}
