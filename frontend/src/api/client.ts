// src/api/client.ts
import axios from 'axios';
import { API_BASE_URL } from '@/utils/config';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

export default apiClient;