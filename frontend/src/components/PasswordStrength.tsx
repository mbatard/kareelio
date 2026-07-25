import { useTranslation } from 'react-i18next';
import { checkPasswordRules } from '../utils/password';

interface PasswordStrengthProps {
  password: string;
  show?: boolean;
}

export function PasswordStrength({ password, show = true }: PasswordStrengthProps) {
  const { t } = useTranslation();
  if (!show || !password) return null;

  const rules = checkPasswordRules(password);

  return (
    <ul className="mt-1 space-y-0.5">
      {rules.map((r) => (
        <li key={r.key} className={`text-xs flex items-center gap-1 ${r.ok ? 'text-green-600 dark:text-green-400' : 'text-gray-400 dark:text-gray-500'}`}>
          {r.ok ? (
            <svg className="w-3 h-3 shrink-0" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" /></svg>
          ) : (
            <svg className="w-3 h-3 shrink-0" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" /></svg>
          )}
          {t(`auth.${r.key}`)}
        </li>
      ))}
    </ul>
  );
}
