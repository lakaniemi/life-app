import { useQuery } from "@tanstack/react-query";
import { z } from "zod";

import { apiFetch } from "@/api/client";

// Enum values the API may add later parse to "unknown" instead of failing the
// whole response.
export const healthSchema = z.object({
  status: z.enum(["ok", "degraded", "unknown"]).catch("unknown"),
  database: z.enum(["ok", "unavailable", "unknown"]).catch("unknown"),
  // The API build's git commit.
  version: z.string(),
});

export type Health = z.infer<typeof healthSchema>;

export const homeKeys = {
  health: () => ["health"] as const,
};

export const useHealth = () =>
  useQuery({
    queryKey: homeKeys.health(),
    queryFn: ({ signal }) => apiFetch("/health", healthSchema, { signal }),
  });
