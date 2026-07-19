import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import fr from './fr.json';
import en from './en.json';

function getBrowserLang(): string {
  const lang = navigator.language.split('-')[0];
  return lang === 'fr' ? 'fr' : 'en';
}

function getStoredLang(): string {
  const stored = localStorage.getItem('kareelio_lang');
  if (stored && ['fr', 'en'].includes(stored)) return stored;
  return getBrowserLang();
}

i18n.use(initReactI18next).init({
  resources: {
    fr: fr,
    en: en,
  },
  lng: getStoredLang(),
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
