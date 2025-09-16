export interface User {
  id: number;
  email: string;
  name?: string;
  role: 'Developer' | 'Engineer' | 'Admin';
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<LoginResponse>;
  logout: () => void;
  isAuthenticated: boolean;
  loading: boolean;
  isDeveloper: boolean;
}
