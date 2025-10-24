module.exports = {
  extends: ["next/core-web-vitals", "../../.eslintrc.js"],
  env: {
    browser: true,
    node: true,
    es2022: true,
  },
  parserOptions: {
    project: "./tsconfig.json",
  },
  rules: {
    // Next.js specific rules
    "@next/next/no-html-link-for-pages": "error",
    "@next/next/no-img-element": "warn",

    // React specific rules
    "react/react-in-jsx-scope": "off",
    "react/prop-types": "off",
    "react-hooks/exhaustive-deps": "warn",

    // Allow console in development
    "no-console": process.env.NODE_ENV === "production" ? "error" : "warn",
  },
  overrides: [
    {
      files: ["**/*.tsx", "**/*.jsx"],
      rules: {
        "@typescript-eslint/explicit-function-return-type": "off",
      },
    },
  ],
};
