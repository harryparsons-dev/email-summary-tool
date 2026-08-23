import { apiRequest } from '../lib/api'
import type { AuthCredentials, AuthResponse } from '../models/auth'

export function signup(credentials: AuthCredentials): Promise<AuthResponse> {
  return apiRequest<AuthResponse, AuthCredentials>('/signup', {
    method: 'POST',
    body: credentials,
  })
}

export function login(credentials: AuthCredentials): Promise<AuthResponse> {
  return apiRequest<AuthResponse, AuthCredentials>('/login', {
    method: 'POST',
    body: credentials,
  })
}

export function logout(): Promise<AuthResponse> {
  return apiRequest<AuthResponse>('/logout', { method: 'POST' })
}

export function refresh(): Promise<AuthResponse> {
  return apiRequest<AuthResponse>('/refresh')
}
