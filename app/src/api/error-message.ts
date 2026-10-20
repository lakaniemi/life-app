import type { ParseKeys } from "i18next";

import { ApiError, NetworkError } from "@/api/errors";
import { i18n } from "@/i18n/i18n";

// Copy for the API error codes that any request can return. Codes specific to
// a feature (e.g. sole_admin) get their copy where that feature handles them.
const messageKeys: Record<string, ParseKeys> = {
  internal: "errors.internal",
};

// User-facing copy for a failed request, picked by the error's code. The
// API's own message is for developers and is never shown.
export const errorMessage = (error: unknown): string => {
  if (error instanceof NetworkError) {
    return i18n.t("errors.network");
  }
  if (error instanceof ApiError && Object.hasOwn(messageKeys, error.code)) {
    return i18n.t(messageKeys[error.code]);
  }
  return i18n.t("errors.generic");
};
