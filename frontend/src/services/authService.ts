import { readonly, ref } from 'vue'
import { apiRequest } from '../lib/api'
import type { AuthCredentials, AuthResponse } from '../models/auth'
import type { User } from '../models/user'

const authenticated = ref(false)
const user = ref<User | null>(null)
let authenticationCheck: Promise<void> | undefined

export const isAuthenticated = readonly(authenticated)

// The signed-in user, set whenever the session is refreshed.
export const currentUser = readonly(user)

export function restoreAuthentication(): Promise<void> {
  authenticationCheck ??= refresh().then(
    () => undefined,
    () => undefined,
  )

  return authenticationCheck
}

export function signup(credentials: AuthCredentials): Promise<AuthResponse> {
  return apiRequest<AuthResponse, AuthCredentials>('/signup', {
    method: 'POST',
    skipSessionRefresh: true,
    body: credentials,
  })
}

export async function login(credentials: AuthCredentials): Promise<AuthResponse> {
  const response = await apiRequest<AuthResponse, AuthCredentials>('/login', {
    method: 'POST',
    skipSessionRefresh: true,
    body: credentials,
  })

  // Login only sets the cookies; refresh loads the signed-in user.
  await refresh()
  return response
}

export async function logout(): Promise<AuthResponse> {
  const response = await apiRequest<AuthResponse>('/logout', {
    method: 'POST',
    skipSessionRefresh: true,
  })

  authenticated.value = false
  user.value = null
  return response
}

export async function refresh(): Promise<User> {
  try {
    const response = await apiRequest<User>('/refresh', { skipSessionRefresh: true })
    authenticated.value = true
    user.value = response
    return response
  } catch (error) {
    authenticated.value = false
    user.value = null
    throw error
  }
}
