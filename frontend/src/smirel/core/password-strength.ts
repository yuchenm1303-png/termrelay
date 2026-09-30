/* A lightweight, local-only *estimate* for explaining password choices.
 * This is not a proof of entropy or an authentication requirement: the backend
 * remains the authority, and the typed password is never sent for scoring. */
export type PasswordStrengthLevel = 0 | 1 | 2 | 3 | 4

export interface PasswordStrength {
  level: PasswordStrengthLevel
  label: string
  hint: string
}

const commonPattern = /password|p@ssw0rd|qwerty|asdfgh|123456|012345|654321|abcdef|letmein|welcome|iloveyou|admin123|muxway/iu
const longRepeat = /(.)\1{3,}/u

export function estimatePasswordStrength(value: string, email = ''): PasswordStrength {
  if (!value) return { level: 0, label: '未输入', hint: '建议使用至少 12 位或更长的密码短语' }

  const length = Array.from(value).length
  const lower = value.toLowerCase()
  const emailLocal = email.trim().split('@')[0].toLowerCase()
  const risky =
    !value.trim()
    || commonPattern.test(lower)
    || longRepeat.test(value)
    || (emailLocal.length >= 4 && lower.includes(emailLocal))

  if (risky) {
    return { level: 1, label: '较弱', hint: '避免常见密码、连续字符或邮箱信息' }
  }

  const variety = [
    /\p{Ll}/u.test(value),
    /\p{Lu}/u.test(value),
    /\p{N}/u.test(value),
    /[^\p{L}\p{N}\s]/u.test(value),
  ].filter(Boolean).length

  // Long, multi-word phrases can be good without forcing symbolic complexity.
  const words = value.trim().split(/\s+/u)
  const passphrase = length >= 20 && words.length >= 3 && new Set(words.map(w => w.toLowerCase())).size >= 3
  let level: PasswordStrengthLevel = 1
  if (length >= 8) level = 2
  if (length >= 12 && variety >= 2) level = 3
  if ((length >= 12 && variety >= 3) || (length >= 16 && variety >= 2) || passphrase) level = 4

  if (level === 4) return { level, label: '较强', hint: '组合不错，请勿在其他网站重复使用' }
  if (level === 3) return { level, label: '良好', hint: '适当增加长度还能进一步提高安全性' }
  if (length < 12) return { level, label: level === 1 ? '较弱' : '一般', hint: '再增加到 12 位或更长会更好' }
  return { level, label: '一般', hint: '可加入数字、符号或使用更长的密码短语' }
}
