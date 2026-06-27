import zhCN from './zh-CN'
import enUS from './en-US'

type Locale = typeof zhCN
const messages: Record<string, Locale> = { 'zh-CN': zhCN, 'en-US': enUS }

let currentLocale = localStorage.getItem('locale') || 'zh-CN'
let _t = messages[currentLocale] || zhCN

export function setLocale(locale: string) {
  if (messages[locale]) {
    currentLocale = locale
    _t = messages[locale]
    localStorage.setItem('locale', locale)
  }
}

export function getLocale() { return currentLocale }

export function t(key: string): string {
  const keys = key.split('.')
  let result: any = _t
  for (const k of keys) {
    result = result?.[k]
  }
  return result || key
}

export function useI18n() {
  return { t, setLocale, getLocale, locale: currentLocale }
}
