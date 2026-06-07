export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: ApiError }

export interface ApiError {
  message: string
}

export function ok<T>(data: T): ApiResult<T> {
  return { ok: true, data }
}

export function err(error: unknown): ApiResult<never> {
  const message = error instanceof Error ? error.message : '调用本地能力失败'
  return { ok: false, error: { message } }
}
