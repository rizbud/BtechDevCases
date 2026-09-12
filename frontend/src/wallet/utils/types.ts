export interface Transaction {
  id: string;
  sender_id?: string;
  sender_email?: string;
  recipient_id: string;
  recipient_email: string;
  amount: number;
  notes?: string;
  created_at: string;
}

export interface PaginationMeta {
  total_records: number;
  total_pages: number;
  current_page: number;
  page_size: number;
}

export interface TransactionFilters {
  start_date: string;
  end_date: string;
  page: number;
  page_size: number;
}
