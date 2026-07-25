export function checkPasswordRules(password: string): { key: string; ok: boolean }[] {
  return [
    { key: 'passwordRuleLength', ok: password.length >= 12 },
    { key: 'passwordRuleUppercase', ok: /[A-Z]/.test(password) },
    { key: 'passwordRuleLowercase', ok: /[a-z]/.test(password) },
    { key: 'passwordRuleDigit', ok: /[0-9]/.test(password) },
    { key: 'passwordRuleSpecial', ok: /[^A-Za-z0-9]/.test(password) },
  ];
}

export function isPasswordValid(password: string): boolean {
  return checkPasswordRules(password).every((r) => r.ok);
}
