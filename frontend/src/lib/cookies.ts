const USER_ID_COOKIE = 'user_id'
const MAX_AGE_DAYS = 365

export function getUserId(): number {
  const match = document.cookie.match(new RegExp(`(?:^|; )${USER_ID_COOKIE}=([^;]*)`))
  const value = match ? decodeURIComponent(match[1]) : null
  if (value !== null) {
    const n = parseInt(value, 10)
    if (!Number.isNaN(n) && n >= 0) return n
  }
  const newId = Math.floor(Math.random() * 1_000_000) + 1
  setUserId(newId)
  return newId
}

export function setUserId(id: number): void {
  const expires = new Date()
  expires.setDate(expires.getDate() + MAX_AGE_DAYS)
  document.cookie = `${USER_ID_COOKIE}=${id}; path=/; max-age=${MAX_AGE_DAYS * 24 * 60 * 60}; SameSite=Lax`
}

export function resetUserId(): number {
  const newId = Math.floor(Math.random() * 1_000_000) + 1
  setUserId(newId)
  return newId
}
