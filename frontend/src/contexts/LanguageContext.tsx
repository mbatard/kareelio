import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import i18n from '../i18n';

type Lang = 'fr' | 'en' | 'system';

interface LanguageContextType {
  language: Lang;
  resolvedLanguage: 'fr' | 'en';
  setLanguage: (l: Lang) => void;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

function getBrowserLang(): 'fr' | 'en' {
  const lang = navigator.language.split('-')[0];
  return lang === 'fr' ? 'fr' : 'en';
}

function getStoredLang(): Lang {
  const stored = localStorage.getItem('kareelio_lang') as Lang | null;
  if (stored && ['fr', 'en', 'system'].includes(stored)) return stored;
  return 'system';
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Lang>(getStoredLang);
  const [browserLang] = useState<'fr' | 'en'>(getBrowserLang);

  useEffect(() => {
    const resolved = language === 'system' ? browserLang : language;
    i18n.changeLanguage(resolved);
  }, [language, browserLang]);

  const setLanguage = (l: Lang) => {
    setLanguageState(l);
    localStorage.setItem('kareelio_lang', l);
  };

  const resolvedLanguage = language === 'system' ? browserLang : language;

  return (
    <LanguageContext.Provider value={{ language, resolvedLanguage, setLanguage }}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLanguage(): LanguageContextType {
  const ctx = useContext(LanguageContext);
  if (!ctx) throw new Error('useLanguage must be used within LanguageProvider');
  return ctx;
}
