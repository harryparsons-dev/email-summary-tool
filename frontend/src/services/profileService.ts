import { apiRequest } from '../lib/api'
import type { User } from '../models/user'

export function getProfile(): Promise<User> {
  return apiRequest<User>('/profile')
}
