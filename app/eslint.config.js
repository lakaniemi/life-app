// https://docs.expo.dev/guides/using-eslint/
const { defineConfig } = require('eslint/config');
const expoConfig = require("eslint-config-expo/flat");

module.exports = defineConfig([
  expoConfig,
  {
    ignores: ["dist/*"],
  },
  {
    rules: {
      "func-style": ["error", "expression"],
      // An anonymous default-exported component loses its state on every Fast
      // Refresh, so route files name it and export it on a separate line.
      "import/no-anonymous-default-export": "error",
      "no-restricted-imports": [
        "error",
        {
          paths: [
            {
              name: "@tanstack/react-query",
              importNames: ["useQuery"],
              message:
                "Use useApiQuery from @/api/use-api-query, which pauses queries of screens out of focus.",
            },
          ],
        },
      ],
    },
  },
  {
    // The wrapper itself.
    files: ["src/api/use-api-query.ts"],
    rules: {
      "no-restricted-imports": "off",
    },
  },
]);
