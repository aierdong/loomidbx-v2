import { describe, expect, it } from "vitest";

import { createBootstrapClient } from "./bootstrapClient";

const readyStatus = {
  name: "LoomiDBX",
  version: "0.1.0",
  runtime: "go",
  ready: true,
  message: "ready",
};

describe("createBootstrapClient", () => {
  it("wraps deterministic binding success as an ok ApiResult", async () => {
    const client = createBootstrapClient({
      BootstrapStatus: () => readyStatus,
    });

    const result = await client.getStatus();

    expect(result).toEqual({ ok: true, data: readyStatus });
  });

  it("wraps deterministic binding failure as an error ApiResult", async () => {
    const client = createBootstrapClient({
      BootstrapStatus: () => {
        throw {
          message: "bootstrap unavailable",
          code: "BOOTSTRAP_UNAVAILABLE",
        };
      },
    });

    const result = await client.getStatus();

    expect(result).toEqual({
      ok: false,
      error: {
        message: "bootstrap unavailable",
        code: "BOOTSTRAP_UNAVAILABLE",
        issues: undefined,
      },
    });
  });
});
