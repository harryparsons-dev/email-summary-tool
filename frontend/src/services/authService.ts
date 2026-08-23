import { readonly, ref } from 'vue'
import { apiRequest } from '../lib/api'
import type { AuthCredentials, AuthResponse } from '../models/auth'

const authenticated = ref(false)
let authenticationCheck: Promise<void> | undefined

export const isAuthenticated = readonly(authenticated)

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
    body: credentials,
  })
}

export async function login(credentials: AuthCredentials): Promise<AuthResponse> {
  const response = await apiRequest<AuthResponse, AuthCredentials>('/login', {
    method: 'POST',
    body: credentials,
  })

  authenticated.value = true
  return response
}

export async function logout(): Promise<AuthResponse> {
  const response = await apiRequest<AuthResponse>('/logout', { method: 'POST' })

  authenticated.value = false
  return response
}

export async function refresh(): Promise<AuthResponse> {
  try {
    const response = await apiRequest<AuthResponse>('/refresh')
    authenticated.value = true
    return response
  } catch (error) {
    authenticated.value = false
    throw error
  }
}
