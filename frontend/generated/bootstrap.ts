import type { BootstrapStatus as BootstrapStatusDto } from "@/types/bootstrap";

export async function BootstrapStatus(): Promise<BootstrapStatusDto> {
  return {
    name: "LoomiDBX",
    version: "0.1.0",
    runtime: "go",
    ready: true,
    message: "LoomiDBX bootstrap skeleton is ready.",
  };
}
