import assert from 'node:assert/strict'
import { test } from 'node:test'

import { convertDetectedLanguage } from './languages'

test('keeps normalized traditional Chinese codes stable', () => {
  assert.equal(convertDetectedLanguage('zhTW'), 'zhTW')
  assert.equal(convertDetectedLanguage('zhtw'), 'zhTW')
})

test('maps browser Chinese locales to interface codes', () => {
  assert.equal(convertDetectedLanguage('zh-TW'), 'zhTW')
  assert.equal(convertDetectedLanguage('zh-Hant-TW'), 'zhTW')
  assert.equal(convertDetectedLanguage('zh-CN'), 'zhCN')
  assert.equal(convertDetectedLanguage('zh'), 'zhCN')
})

test('leaves non-Chinese locales unchanged', () => {
  assert.equal(convertDetectedLanguage('en'), 'en')
  assert.equal(convertDetectedLanguage('fr-FR'), 'fr-FR')
})
