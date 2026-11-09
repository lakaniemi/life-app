import { z } from "zod";

import { apiUrl } from "@/api/config";
import { ApiError, NetworkError } from "@/api/errors";

// React Native's fetch has no timeout of its own, so on a flaky connection a
// request could otherwise hang until the OS gives up.
const requestTimeoutMs = 15_000;

const errorEnvelopeSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
    fields: z
      .array(z.object({ field: z.string(), rule: z.string() }))
      .optional(),
  }),
});

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  // Sent as JSON.
  body?: unknown;
  // From TanStack Query, which aborts the request when the query is cancelled.
  signal?: AbortSignal;
};

// Requests `path` from the API and parses the response with `schema`, so a
// response that doesn't match fails here, loudly, rather than as `undefined`
// deep in a screen. Throws ApiError for an error status and NetworkError when
// there's no response. A 204 is parsed as `undefined` (use `z.void()`).
export const apiFetch = async <T>(
  path: string,
  schema: z.ZodType<T>,
  options: RequestOptions = {},
): Promise<T> => {
  const response = await send(path, options);

  if (!response.ok) {
    throw await toApiError(response);
  }
  if (response.status === 204) {
    return schema.parse(undefined);
  }
  return schema.parse(await readJson(response, options.method ?? "GET", path));
};

// The API always answers in JSON, so a body that isn't means the request
// never reached it: typically a captive portal's login page.
const readJson = async (
  response: Response,
  method: string,
  path: string,
): Promise<unknown> => {
  try {
    return await response.json();
  } catch {
    throw new NetworkError(`${method} ${path} got a response that isn't JSON`);
  }
};

const send = async (
  path: string,
  { method = "GET", body, signal }: RequestOptions,
): Promise<Response> => {
  // One controller aborts the request either on timeout or when the caller's
  // signal aborts.
  const controller = new AbortController();
  const abort = () => controller.abort();
  const timeout = setTimeout(abort, requestTimeoutMs);
  if (signal?.aborted) {
    abort();
  }
  signal?.addEventListener("abort", abort);

  const headers: Record<string, string> = { Accept: "application/json" };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  try {
    return await fetch(`${apiUrl}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    });
  } catch (error) {
    // A cancellation by the caller is passed on as it is, for TanStack Query
    // to recognize. Anything else means no response.
    if (signal?.aborted) {
      throw error;
    }
    if (controller.signal.aborted) {
      throw new NetworkError(`${method} ${path} timed out`, {
        timedOut: true,
      });
    }
    throw new NetworkError(
      `${method} ${path} got no response: ${String(error)}`,
    );
  } finally {
    clearTimeout(timeout);
    signal?.removeEventListener("abort", abort);
  }
};

const toApiError = async (response: Response): Promise<ApiError> => {
  const json: unknown = await response.json().catch(() => undefined);
  const envelope = errorEnvelopeSchema.safeParse(json);
  if (!envelope.success) {
    // E.g. a proxy's HTML error page in front of the API.
    return new ApiError(
      response.status,
      "unknown",
      `HTTP ${response.status} without an error envelope`,
    );
  }
  const { code, message, fields } = envelope.data.error;
  return new ApiError(response.status, code, message, fields);
};
