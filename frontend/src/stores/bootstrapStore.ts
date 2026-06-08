import { reactive } from "vue";
import { bootstrapClient } from "@/api/bootstrapClient";

type BootstrapState = "loading" | "ready" | "error";

export interface BootstrapViewState {
  state: BootstrapState;
  title: string;
  message: string;
}

const bootstrapState = reactive<BootstrapViewState>({
  state: "loading",
  title: "Loading bootstrap status",
  message: "正在验证前后端骨架调用链路。",
});

export function useBootstrapStore(): BootstrapViewState {
  return bootstrapState;
}

export async function loadBootstrapStatus(): Promise<void> {
  bootstrapState.state = "loading";
  const result = await bootstrapClient.getStatus();
  if (result.ok) {
    bootstrapState.state = result.data.ready ? "ready" : "error";
    bootstrapState.title = `${result.data.name} ${result.data.version}`;
    bootstrapState.message = result.data.message;
    return;
  }

  bootstrapState.state = "error";
  bootstrapState.title = "Bootstrap status failed";
  bootstrapState.message = result.error.message;
}
