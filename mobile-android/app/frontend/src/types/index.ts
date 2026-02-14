export interface User {
  id: number;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: string;
  status: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface Account {
  id: number;
  user_id: number;
  account_number: string;
  account_type: string;
  balance: string;
  currency: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Transfer {
  id: number;
  reference_id: string;
  from_account_id: number;
  to_account_id: number;
  amount: string;
  currency: string;
  status: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface Payment {
  id: number;
  reference_id: string;
  account_id: number;
  user_id: number;
  payment_type: string;
  recipient_name?: string;
  recipient_account?: string;
  recipient_bank?: string;
  amount: string;
  currency: string;
  description?: string;
  status: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;
  processed_at?: string;
}

export interface Notification {
  id: number;
  user_id: number;
  type: string;
  channel: string;
  title: string;
  content: string;
  metadata?: Record<string, unknown>;
  status: string;
  read_at?: string;
  sent_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAccountRequest {
  user_id: number;
  account_type: string;
  currency?: string;
}

export interface CreateTransferRequest {
  from_account_id: number;
  to_account_id: number;
  amount: string;
  currency?: string;
}

export interface CreatePaymentRequest {
  account_id: number;
  payment_type: string;
  recipient_name?: string;
  recipient_account?: string;
  recipient_bank?: string;
  amount: string;
  currency?: string;
  description?: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: {
    id: number;
    username: string;
    role: string;
  };
}

export interface AccountListResponse {
  accounts: Account[];
  total: number;
}

export interface TransferListResponse {
  transfers: Transfer[];
  total: number;
}

export interface PaymentListResponse {
  payments: Payment[];
  total: number;
}

export interface NotificationListResponse {
  notifications: Notification[];
  total: number;
  unread: number;
}

export interface UserListResponse {
  users: User[];
  total: number;
}

export interface AuthUser {
  id: number;
  username: string;
  role: string;
}

export interface Card {
  id: number;
  account_id: number;
  card_number: string;
  card_type: string;
  cardholder_name: string;
  expiration_month: number;
  expiration_year: number;
  status: string;
  daily_limit: string;
  monthly_limit: string;
  per_transaction_limit: string;
  daily_used: string;
  monthly_used: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCardRequest {
  account_id: number;
  card_type: string;
  cardholder_name: string;
}

export interface UpdateCardRequest {
  status?: string;
  daily_limit?: string;
  monthly_limit?: string;
  per_transaction_limit?: string;
}

export interface CardListResponse {
  cards: Card[];
  total: number;
}
