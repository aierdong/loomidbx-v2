export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: ApiError };

export interface ApiError {
  message: string;
  code?: string;
  issues?: ApiIssue[];
}

export interface ApiIssue {
  path: string;
  code: string;
  severity: string;
  message: string;
}

export function ok<T>(data: T): ApiResult<T> {
  return { ok: true, data };
}

export function err(error: unknown): ApiResult<never> {
  if (isApiErrorLike(error)) {
    return {
      ok: false,
      error: {
        message: error.message,
        code: error.code,
        issues: error.issues,
      },
    };
  }

  const message = error instanceof Error ? error.message : "调用本地能力失败";
  return { ok: false, error: { message } };
}

function isApiErrorLike(error: unknown): error is ApiError {
  if (!isRecord(error) || typeof error.message !== "string") {
    return false;
  }

  return (
    (error.code === undefined || typeof error.code === "string") &&
    (error.issues === undefined || isApiIssueList(error.issues))
  );
}

function isApiIssueList(value: unknown): value is ApiIssue[] {
  return Array.isArray(value) && value.every(isApiIssue);
}

function isApiIssue(value: unknown): value is ApiIssue {
  return (
    isRecord(value) &&
    typeof value.path === "string" &&
    typeof value.code === "string" &&
    typeof value.severity === "string" &&
    typeof value.message === "string"
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
