import { z } from "zod";

import { apiFetch } from "@/api/client";
import { appCompatLevel, markAppOutdated } from "@/api/compat";
import { ApiError, NetworkError } from "@/api/errors";

jest.mock("@/api/compat", () => ({
  ...jest.requireActual<typeof import("@/api/compat")>("@/api/compat"),
  markAppOutdated: jest.fn(),
}));

// jest.setup.ts mocks the API URL as https://api.test.
const fetchMock = jest.spyOn(globalThis, "fetch");

const jsonResponse = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });

const itemSchema = z.object({ id: z.string() });

afterEach(() => {
  fetchMock.mockReset();
  jest.mocked(markAppOutdated).mockClear();
  // Here rather than in the timeout test, so that a failure there doesn't
  // leave fake timers on for the tests after it.
  jest.useRealTimers();
});

test("requests the path from the API and parses the response", async () => {
  fetchMock.mockResolvedValue(jsonResponse(200, { id: "a", extra: true }));

  await expect(apiFetch("/items/a", itemSchema)).resolves.toEqual({ id: "a" });
  expect(fetchMock).toHaveBeenCalledWith(
    "https://api.test/items/a",
    expect.objectContaining({ method: "GET" }),
  );
});

test("sends the compatibility level with every request", async () => {
  fetchMock.mockResolvedValue(jsonResponse(200, { id: "a" }));

  await apiFetch("/items/a", itemSchema);

  expect(fetchMock).toHaveBeenCalledWith(
    "https://api.test/items/a",
    expect.objectContaining({
      headers: expect.objectContaining({
        "App-Compat-Level": String(appCompatLevel),
      }),
    }),
  );
});

test("marks the app outdated when the API turns it away", async () => {
  fetchMock.mockResolvedValue(
    jsonResponse(400, {
      error: { code: "app_outdated", message: "this app version is too old" },
    }),
  );

  await expect(apiFetch("/items", itemSchema)).rejects.toMatchObject({
    status: 400,
    code: "app_outdated",
  });
  expect(markAppOutdated).toHaveBeenCalledTimes(1);
});

test("doesn't mark the app outdated for other errors", async () => {
  fetchMock.mockResolvedValue(
    jsonResponse(400, {
      error: { code: "invalid_body", message: "invalid request body" },
    }),
  );

  await expect(apiFetch("/items", itemSchema)).rejects.toBeInstanceOf(ApiError);
  expect(markAppOutdated).not.toHaveBeenCalled();
});

test("sends a body as JSON", async () => {
  fetchMock.mockResolvedValue(jsonResponse(201, { id: "a" }));

  await apiFetch("/items", itemSchema, { method: "POST", body: { name: "x" } });

  expect(fetchMock).toHaveBeenCalledWith(
    "https://api.test/items",
    expect.objectContaining({
      method: "POST",
      body: '{"name":"x"}',
      headers: expect.objectContaining({ "Content-Type": "application/json" }),
    }),
  );
});

test("parses a 204 as undefined", async () => {
  fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

  await expect(
    apiFetch("/items/a", z.void(), { method: "DELETE" }),
  ).resolves.toBeUndefined();
});

test("turns the error envelope into an ApiError", async () => {
  fetchMock.mockResolvedValue(
    jsonResponse(400, {
      error: {
        code: "invalid_body",
        message: "invalid request body",
        fields: [{ field: "name", rule: "required" }],
      },
    }),
  );

  const error = await apiFetch("/items", itemSchema).catch((e: unknown) => e);

  expect(error).toBeInstanceOf(ApiError);
  expect(error).toMatchObject({
    status: 400,
    code: "invalid_body",
    fields: [{ field: "name", rule: "required" }],
  });
});

test("gives an error status without an envelope the code unknown", async () => {
  fetchMock.mockResolvedValue(
    new Response("<html>Bad Gateway</html>", { status: 502 }),
  );

  await expect(apiFetch("/items", itemSchema)).rejects.toMatchObject({
    status: 502,
    code: "unknown",
  });
});

test("throws a NetworkError when there's no response", async () => {
  fetchMock.mockRejectedValue(new TypeError("Network request failed"));

  const error = await apiFetch("/items", itemSchema).catch((e: unknown) => e);

  expect(error).toBeInstanceOf(NetworkError);
  expect(error).toMatchObject({ timedOut: false });
});

test("throws a NetworkError for a success that isn't JSON", async () => {
  fetchMock.mockResolvedValue(
    new Response("<html>Sign in to the Wi-Fi</html>", { status: 200 }),
  );

  await expect(apiFetch("/items", itemSchema)).rejects.toBeInstanceOf(
    NetworkError,
  );
});

test("throws a NetworkError when the request times out", async () => {
  jest.useFakeTimers();
  fetchMock.mockImplementation(
    (_input, init) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () =>
          reject(new Error("Aborted")),
        );
      }),
  );

  const request = apiFetch("/items", itemSchema);
  jest.advanceTimersByTime(15_000);

  const error = await request.catch((e: unknown) => e);
  expect(error).toBeInstanceOf(NetworkError);
  expect(error).toMatchObject({ timedOut: true });
});

test("passes a cancellation by the caller on as it is", async () => {
  const aborted = new Error("Aborted");
  fetchMock.mockImplementation(
    (_input, init) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () => reject(aborted));
      }),
  );
  const controller = new AbortController();

  const request = apiFetch("/items", itemSchema, { signal: controller.signal });
  controller.abort();

  await expect(request).rejects.toBe(aborted);
});

test("throws when the response doesn't match the schema", async () => {
  fetchMock.mockResolvedValue(jsonResponse(200, { id: 1 }));

  await expect(apiFetch("/items/a", itemSchema)).rejects.toBeInstanceOf(
    z.ZodError,
  );
});
