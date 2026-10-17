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
    },
  },
]);
