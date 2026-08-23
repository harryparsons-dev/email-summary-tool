import type { User } from './user'

export type AuthCredentials = Pick<User, 'email'> & {
  password: string
}

export type AuthResponse = string
