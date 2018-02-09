module.exports = {
  parser: "babel-eslint",
  extends: [
    "eslint:recommended",
    "airbnb",
    "plugin:react/recommended",
    "prettier",
    "prettier/react",
  ],
  plugins: ["react", "prettier"],
  rules: {
    "import/prefer-default-export": ["off"],
    "prettier/prettier": [
      "error",
      {
        trailingComma: "all",
      },
    ],
  },
  env: {
    es6: true,
    node: true,
  },
};
