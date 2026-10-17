import { errorMessage } from "@/api/error-message";
import { ApiError, NetworkError } from "@/api/errors";

test("says it couldn't connect when there was no response", () => {
  expect(errorMessage(new NetworkError("timed out"))).toBe("Couldn't connect.");
});

test("picks the copy for a known code", () => {
  expect(
    errorMessage(new ApiError(500, "internal", "internal server error")),
  ).toBe("Something went wrong on the server.");
});

test("falls back to a generic message for an unknown code", () => {
  expect(errorMessage(new ApiError(409, "something_new", "new rule"))).toBe(
    "Something went wrong.",
  );
});

test("falls back to a generic message for other errors", () => {
  expect(errorMessage(new Error("boom"))).toBe("Something went wrong.");
});
