import { pickLanguage } from "@/i18n/i18n";

test("picks the first preferred language the app has copy for", () => {
  expect(pickLanguage([{ languageCode: "fi" }, { languageCode: "en" }])).toBe(
    "en",
  );
});

test("falls back to English when no preferred language is supported", () => {
  expect(pickLanguage([{ languageCode: "fi" }, { languageCode: "sv" }])).toBe(
    "en",
  );
});

test("skips locales without a language code", () => {
  expect(pickLanguage([{ languageCode: null }])).toBe("en");
});
