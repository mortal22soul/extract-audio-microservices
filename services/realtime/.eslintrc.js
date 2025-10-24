module.exports = {
  extends: ["../../.eslintrc.js"],
  env: {
    node: true,
    es2022: true,
  },
  parserOptions: {
    project: "./tsconfig.json",
  },
  rules: {
    // Node.js specific rules
    "no-console": "warn", // Allow console in server-side code
    "@typescript-eslint/no-var-requires": "off", // Allow require() for dynamic imports

    // Socket.IO specific
    "@typescript-eslint/no-explicit-any": "warn", // Socket.IO types can be complex
  },
};
