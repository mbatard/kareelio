module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [2, 'always', [
      'feat', 'fix', 'chore', 'refactor', 'test', 'docs', 'ci', 'perf', 'style', 'build',
    ]],
    'scope-enum': [1, 'always', [
      'auth', 'admin', 'profile', 'jobs', 'k8s', 'docker', 'migration', 'i18n', 'deps', 'ci',
    ]],
    'subject-full-stop': [2, 'never', '.'],
  },
};
