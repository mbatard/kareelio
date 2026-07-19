export function isValidEmail(email: string): boolean {
  email = email.trim();

  if (email.length === 0 || email.length > 254) {
    return false;
  }

  if (/[\x00-\x1F\x7F]/.test(email)) {
    return false;
  }

  const atIdx = email.lastIndexOf('@');
  if (atIdx < 1) {
    return false;
  }

  const local = email.slice(0, atIdx);
  const domain = email.slice(atIdx + 1);

  if (local.length === 0 || local.length > 64) {
    return false;
  }
  if (domain.length === 0) {
    return false;
  }

  if (!isValidLocal(local)) {
    return false;
  }

  if (!isValidDomain(domain)) {
    return false;
  }

  return true;
}

function isValidLocal(local: string): boolean {
  if (local[0] === '.' || local[local.length - 1] === '.') {
    return false;
  }

  for (let i = 0; i < local.length; i++) {
    const c = local[i];
    if (
      !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
        c === '.' || c === '_' || c === '+' || c === '-')
    ) {
      return false;
    }
  }

  return true;
}

function isValidDomain(domain: string): boolean {
  if (domain.startsWith('.') || domain.endsWith('.')) {
    return false;
  }

  const labels = domain.split('.');
  if (labels.length < 2) {
    return false;
  }

  for (const label of labels) {
    if (label.length === 0 || label.length > 63) {
      return false;
    }
    if (label[0] === '-' || label[label.length - 1] === '-') {
      return false;
    }
    for (let i = 0; i < label.length; i++) {
      const c = label[i];
      if (
        !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c === '-')
      ) {
        return false;
      }
    }
  }

  const tld = labels[labels.length - 1];
  if (tld.length < 2) {
    return false;
  }
  if (tld.startsWith('xn--') && tld.length < 5) {
    return false;
  }

  return true;
}
