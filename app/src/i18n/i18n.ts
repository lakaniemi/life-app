// Hermes doesn't implement Intl.PluralRules, which i18next needs for plural
// keys (`_one`, `_other`). This installs a polyfill only when it's missing.
import "intl-pluralrules";

import { getLocales, type Locale } from "expo-localization";
import { createInstance } from "i18next";
import { initReactI18next } from "react-i18next";

import en from "@/i18n/en.json";

export const resources = {
  en: { translation: en },
};

type Language = keyof typeof resources;

const fallbackLanguage: Language = "en";

// The first of the device's preferred languages that the app has copy for.
export function pickLanguage(locales: Pick<Locale, "languageCode">[]): Language {
  for (const { languageCode } of locales) {
    if (languageCode !== null && Object.hasOwn(resources, languageCode)) {
      return languageCode as Language;
    }
  }
  return fallbackLanguage;
}

export const i18n = createInstance();

// initReactI18next hands this instance to useTranslation(). The resources are
// bundled, so init runs synchronously: the first render already has its copy.
void i18n.use(initReactI18next).init({
  resources,
  lng: pickLanguage(getLocales()),
  fallbackLng: fallbackLanguage,
  // React already escapes rendered strings.
  interpolation: { escapeValue: false },
});
