import { client } from './client'

export interface RegisterInput {
  email: string
  username: string
  password: string
}

export interface LoginInput {
  email: string
  password: string
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
}

export const authApi = {
  register: (data: RegisterInput) =>
    client.post('/auth/register', data),

  login: (data: LoginInput) =>
    client.post<AuthTokens>('/auth/login', data),

  verifyEmail: (token: string) =>
    client.get(`/auth/verify-email?token=${token}`),

  forgotPassword: (email: string) =>
    client.post('/auth/forgot-password', { email }),

  resetPassword: (token: string, new_password: string) =>
    client.post('/auth/reset-password', { token, new_password }),
}
