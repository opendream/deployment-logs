import { createContext, useContext, useState, useCallback, useMemo, type ReactNode } from 'react'

interface User {
  id: number
  username: string
  role: string
}

interface AuthContextType {
  token: string | null
  user: User | null
  isAdmin: boolean
  login: (token: string, user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType>(null!)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [user, setUser] = useState<User | null>(() => {
    const s = localStorage.getItem('user')
    return s ? JSON.parse(s) : null
  })

  const login = useCallback((token: string, user: User) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    setToken(token)
    setUser(user)
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setToken(null)
    setUser(null)
  }, [])

  const isAdmin = user?.role === 'admin'

  const value = useMemo(() => ({ token, user, isAdmin, login, logout }), [token, user, isAdmin, login, logout])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  return useContext(AuthContext)
}

export function apiHeaders(token: string | null): HeadersInit {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  return headers
}

export function getBasePath(): string {
  return (window as any).__BASE_PATH__ || ''
}

export function apiUrl(path: string): string {
  return getBasePath() + path
}

export function apiFetch(input: string, init?: RequestInit): Promise<Response> {
  return fetch(input, init).then(resp => {
    if (resp.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = getBasePath() + '/login'
      return Promise.reject(new Error('Unauthorized'))
    }
    return resp
  })
}
