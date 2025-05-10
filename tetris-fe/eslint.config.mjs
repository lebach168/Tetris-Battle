import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  plugins= {
      "react-hooks": reactHooks,
    },
    rules= {
      // Vô hiệu hóa quy tắc cũ
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": "error",

      // Thêm quy tắc `react-hooks`
      "react-hooks/rules-of-hooks": "error",        // Bắt buộc đúng quy tắc dùng hook
      "react-hooks/exhaustive-deps": "warn",        // Cảnh báo thiếu dependencies
    },
  
];

export default eslintConfig;
