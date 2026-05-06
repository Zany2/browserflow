import js from '@eslint/js'
import { globalIgnores } from 'eslint/config'
import globals from 'globals'
import pluginVue from 'eslint-plugin-vue'
import pluginOxlint from 'eslint-plugin-oxlint'
import skipFormatting from 'eslint-config-prettier/flat'

export default [
  {
    name: 'app/files-to-lint',
    files: ['**/*.{js,mjs,cjs,vue}'],
  },
  globalIgnores([
    '**/node_modules/**',
    '**/dist/**',
    '**/dist-ssr/**',
    '**/coverage/**',
  ]),
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  js.configs.recommended,
  ...pluginVue.configs['flat/essential'],
  ...pluginOxlint.buildFromOxlintConfigFile('.oxlintrc.json'),
  {
    rules: {
      // Unused vars 未使用变量，允许下划线开头的占位参数
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },
  skipFormatting,
]
