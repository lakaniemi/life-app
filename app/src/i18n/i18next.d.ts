import "i18next";

import type { resources } from "@/i18n/i18n";

// Types t() keys against the English copy, so a mistyped key fails typecheck.
declare module "i18next" {
  interface CustomTypeOptions {
    resources: (typeof resources)["en"];
  }
}
