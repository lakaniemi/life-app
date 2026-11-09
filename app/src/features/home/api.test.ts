import { healthSchema } from "@/features/home/api";

test("parses a status the app doesn't know yet as unknown", () => {
  expect(
    healthSchema.parse({
      status: "maintenance",
      database: "read_only",
      version: "abc",
    }),
  ).toEqual({ status: "unknown", database: "unknown", version: "abc" });
});
